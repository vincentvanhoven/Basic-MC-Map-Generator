import time
import io
import zlib
import os

from tqdm import tqdm
from Chunk import Chunk
import json
import multiprocessing as mp
from TerminalColors import TerminalColors


# The region file header is 8KiB
#   4KiB chunk locations
#   4KiB chunk timestamps (last updated)
class RegionFilesReader:
    def __init__(self):
        self.chunks = []
        # self.progressBar = None

    def process_region_file(self, filePath):
        return self.processRegionFile(filePath)

    def iterateDirectory(self, directoryPath, ignoreCache):
        if ignoreCache is False:
            if self.attemptLoadCacheFile(directoryPath):
                return

        regionFiles = []

        for subdir, dirs, files in os.walk(directoryPath):
            for file in files:
                if file.endswith((".mca")):
                    regionFiles.append(f'{directoryPath}/{file}')

        # baseline: ~10s
        progressBar = tqdm(total=len(regionFiles))
        pool = mp.Pool(processes=len(regionFiles))

        def callback(chunkArray):
            self.chunks = self.chunks + chunkArray
            progressBar.set_description_str(f'Reading region files ({len(self.chunks)} chunks)')
            progressBar.update(1)

        for file in regionFiles:
            pool.apply_async(self.process_region_file, (file,), callback=callback, error_callback=(lambda x: print(x)))

        # Indicate that no new jobs will be pushed to the pool
        pool.close()
        # Wait for all jobs to finish
        pool.join()

        # Ensure the progress bar is closed properly
        progressBar.close()

        # Writing chunks to JSON cache file
        print("Writing chunks to cache...")
        self.writeCacheFile(directoryPath)

    def attemptLoadCacheFile(self, directoryPath):
        cacheDir = os.path.join(os.path.dirname(__file__), 'cache')

        if os.path.exists(cacheDir) == False:
            os.mkdir(cacheDir)

        for subdir, dirs, files in os.walk(cacheDir):
            for file in files:
                if file.endswith(("_cache.json")):
                    with open(os.path.join(cacheDir, file), 'r') as fileContents:
                        if fileContents != "":
                            JSON = json.load(fileContents)

                            if JSON.get("regionFolderPath") == directoryPath:
                                print(TerminalColors.OKGREEN + "Found a cache file. Loading chunk data from cache file..." + TerminalColors.ENDC)
                                self.chunks = list(map(lambda chunk: Chunk(
                                    chunk.get("dataOffset"),
                                    chunk.get("biome"),
                                    chunk.get("timestamp"),
                                    chunk.get("isLoaded"),
                                    chunk.get("x"),
                                    chunk.get("z")
                                ), JSON.get("chunks")))

                                return True
        return False

    def writeCacheFile(self, directoryPath):
        cacheDirPath = os.path.join(os.path.dirname(__file__), 'cache/')
        cacheFilePath = os.path.join(cacheDirPath, f'{time.time()}_cache.json')

        cacheFile = {
            "regionFolderPath": directoryPath,
            "chunks": list(map(lambda chunk: {
                "dataOffset": chunk.dataOffset,
                "biome": chunk.biome,
                "timestamp": chunk.timestamp,
                "isLoaded": chunk.isLoaded,
                "x": chunk.x,
                "z": chunk.z,
            }, self.chunks))
        }

        if os.path.exists(cacheDirPath) == False:
            os.mkdir(cacheDirPath)

        with open(cacheFilePath, "w") as outfile:
            outfile.write(json.dumps(cacheFile, indent=4))
            outfile.close()

    def processRegionFile(self, filePath):
        # Opening the binary file in binary mode as rb (read binary)
        with open(filePath, mode="rb") as file:
            chunks = []

            # Read the locations of the chunk data in the region file
            for chunkIndex in range(1024):
                # The first 3 bytes define the offset (in bytes) from the start of the region file to the start of the chunk
                # data for this particular chunk
                chunkDataOffset = int.from_bytes(file.read(3), byteorder='big')
                # The following byte indicates the approximate size of the chunk data
                int.from_bytes(file.read(1), byteorder='big')

                # Persist this data in the chunks array
                chunks.append(Chunk(chunkDataOffset))

            # Read the chunk timestamps (last updated)
            for chunkIndex in range(1024):
                timestamp = int.from_bytes(file.read(4), byteorder='big')
                chunks[chunkIndex].timestamp = timestamp

            # Read the chunk x & z position data
            for chunkIndex in range(1024):
                # If the data offset of the chunk is 0, there is no data for this chunk (it hasn't been generated yet)
                if chunks[chunkIndex].dataOffset == 0:
                    # Skip the loading of this chunk's data
                    continue

                # Place the cursor at the start of the current chunk's data
                file.seek(chunks[chunkIndex].dataOffset * 1024 * 4, 0)
                # Get the data length of this chunk
                chunkLength = int.from_bytes(file.read(4), byteorder='big')
                compressionType = int.from_bytes(file.read(1), byteorder='big')

                # Read and decompress the chunk data. The compression type is assumed to be zlib.
                chunkData = zlib.decompress(file.read(chunkLength))
                # Stream the chunk data (NBT)
                chunkDataStream = io.BytesIO(chunkData)

                # Read the opening tag from the NBT stream (should be Compound, value 10)
                firstTag = int.from_bytes(chunkDataStream.read(1), byteorder='big')
                # Assert that this is a valid chunk data set
                if firstTag != 10:
                    raise Exception("Something went wrong. The NBT data should start with a Compound tag.")

                # Find the offset of the chunk's status data
                statusDataOffset = chunkDataStream.read().find(b'Status')
                # Position the cursor to a bit before the status value
                chunkDataStream.seek(statusDataOffset + 1 + 6, 0)
                # Get the length of the status value
                statusValueLength = int.from_bytes(chunkDataStream.read(2), byteorder='big', signed=False)
                # Read the status value
                statusValue = chunkDataStream.read(statusValueLength)

                # If the chunk status is not 'full', it's not fully generated yet
                if statusValue != b'full':
                    # Don't use incomplete chunks
                    continue

                # Reset the cursor
                chunkDataStream.seek(0, 0)

                # Find the offset of the chunk's x position data
                xPosDataOffset = chunkDataStream.read().find(b'xPos')
                # Read over the NBT tag name
                chunkDataStream.seek(xPosDataOffset + 4, 0)
                # Read the chunk's x position data
                chunks[chunkIndex].x = int.from_bytes(chunkDataStream.read(4), byteorder='big', signed=True)

                # Reset the cursor
                chunkDataStream.seek(0, 0)

                # Find the offset of the chunk's z position data
                zPosDataOffset = chunkDataStream.read().find(b'zPos')

                # Read over the NBT tag name
                chunkDataStream.seek(zPosDataOffset + 4, 0)
                # Read the chunk's z position data
                chunks[chunkIndex].z = int.from_bytes(chunkDataStream.read(4), byteorder='big', signed=True)

                # Reset the cursor
                chunkDataStream.seek(0, 0)

                # Find the offset of the 'sections' list
                sectionDataOffset = chunkDataStream.read().find(b'sections')
                # Read over the NBT tag name
                chunkDataStream.seek(sectionDataOffset + 9, 0)
                # Read the 'sections' list length
                sectionsListLength = int.from_bytes(chunkDataStream.read(4), byteorder='big', signed=True)

                # Prepare storing the biomes of this chunk
                chunkBiomes = []

                for sectionsIndex in range(0,  sectionsListLength):
                    # Get the position of the stream cursor
                    currentOffset = chunkDataStream.tell()
                    # Find the index of the 'sections.biomes' list (corrected for the current position of the cursor)
                    biomesOffset = chunkDataStream.read().find(b'biomes') + currentOffset
                    # Place cursor at the 'sections.biomes' opening tag
                    chunkDataStream.seek(biomesOffset)

                    # Get the position of the stream cursor
                    currentOffset = chunkDataStream.tell()
                    # Find the index of the 'sections.biomes.palette' list (corrected for the current position of the cursor)
                    paletteOffset = chunkDataStream.read().find(b'palette') + currentOffset
                    # Place cursor at the 'sections.biomes.palette' data
                    chunkDataStream.seek(paletteOffset + 8)
                    # Read the length of the 'sections.biomes.palette' list (in items, not bytes)
                    paletteListLength = int.from_bytes(chunkDataStream.read(4), byteorder='big', signed=True)

                    # Iterate over all the palettes that should be in the 'sections.biomes.palette' list
                    for paletteIndex in range(0, paletteListLength):
                        # Get the position of the stream cursor
                        currentOffset = chunkDataStream.tell()
                        # Find the index of the current palette item (corrected for the current position of the cursor)
                        paletteIndexOffset = chunkDataStream.read().find(b'minecraft:') + currentOffset
                        # Place cursor just before the current palette item data
                        chunkDataStream.seek(paletteIndexOffset - 2)
                        # Read the length of the current palette item data
                        paletteIndexValueLength = int.from_bytes(chunkDataStream.read(2), byteorder='big', signed=False)
                        # Read the palette item data and store it in 'chunkBiomes' as a UTF-8 string
                        chunkBiomes.append(chunkDataStream.read(paletteIndexValueLength).decode("utf-8"))

                        # if not (len(chunkBiomes) > 0):
                            # print("something went oddly here")

                if len(chunkBiomes) > 0:
                    chunks[chunkIndex].biome = max(set(chunkBiomes), key = chunkBiomes.count)

                # Mark the chunk instance as valid
                chunks[chunkIndex].isLoaded = True

        chunks = [chunk for chunk in chunks if chunk.isLoaded != False]

        # Return chunks
        return chunks
