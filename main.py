from PIL import Image, ImageDraw
from tqdm import tqdm
from BiomeColors import BiomeColors
from RegionFilesReader import RegionFilesReader
import datetime
import sys
import os
import time

from TerminalColors import TerminalColors


if __name__ == '__main__':
    dateTimeStart = datetime.datetime.now()

    sourcePath = sys.argv[1]
    amountOfProcesses = 48
    ignoreCache = False

    if len(sys.argv) > 2:
        amountOfProcesses = int(sys.argv[2]) if sys.argv[2] else 48
    if len(sys.argv) > 3:
        ignoreCache = sys.argv[3] if sys.argv[3] else False

    outputDir = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'output')

    if os.path.exists(outputDir) == False:
        os.mkdir(outputDir)

    outputFilePath = os.path.join(outputDir, f'map_{time.time()}.png')

    regionFilesReader = RegionFilesReader()
    regionFilesReader.iterateDirectory(sourcePath, ignoreCache, amountOfProcesses)

    print(f'Read {len(regionFilesReader.chunks)} chunks from the region files or cache.')

    progressBar = tqdm(total=len(regionFilesReader.chunks))
    progressBar.set_description_str(f'Rendering chunks')

    minX = 0
    maxX = 0
    minZ = 0
    maxZ = 0

    for chunk in regionFilesReader.chunks:
        if chunk.x <= minX:
            minX = chunk.x
        if chunk.x >= maxX:
            maxX = chunk.x
        if chunk.z <= minZ:
            minZ = chunk.z
        if chunk.z >= maxZ:
            maxZ = chunk.z

    # print(f'Chunk range runs from [{minX}, {minZ}] to [{maxX}, {maxZ}]')

    tileSize = 1
    imageWidth = (abs(minX) + maxX) * tileSize
    imageHeight = (abs(minZ) + maxZ) * tileSize
    imageOriginX = abs(minX) * tileSize
    imageOriginZ = abs(minZ) * tileSize

    im = Image.new(mode="RGBA", size=(imageWidth, imageHeight))
    draw = ImageDraw.Draw(im)

    biomeColors = BiomeColors()

    for chunk in regionFilesReader.chunks:
        color = biomeColors.getColor(chunk.biome)

        draw.rectangle(
            (
                imageOriginX + (chunk.x * tileSize),
                imageOriginZ + (chunk.z * tileSize),
                imageOriginX + ((chunk.x + 1) * tileSize),
                imageOriginZ + ((chunk.z + 1) * tileSize),
            ),
            fill=color
        )
        progressBar.update(1)

    im.save(outputFilePath)
    progressBar.close()

    print(f'{TerminalColors.OKGREEN}The file was generated in: {(datetime.datetime.now() - dateTimeStart).total_seconds()} seconds and can be found here:{TerminalColors.ENDC}')
    print(outputFilePath)

