package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/Tnze/go-mc/nbt"
)

const rootDir string = "C:\\Users\\Vincent\\Desktop\\MC Server\\Anarchy\\world\\region\\"

func int24BinaryToInt32(bytes []byte) uint32 {
	return binary.BigEndian.Uint32(append([]byte{0x00}, bytes...))
}

func getIntParamWithFallback(r *http.Request, paramName string, fallback int) int {
	urlParam := r.URL.Query().Get(paramName)

	if len(urlParam) > 0 {
		// parsedInt, error := strconv.Atoi(urlParam)
		parsedInt, error := strconv.Atoi(urlParam)

		TEST_GITHUB_ACTION_FAIL

		if error == nil {
			return parsedInt
		} else {
			fmt.Println(error)
		}
	}

	return fallback
}

func findByteStringAllInReadSeeker(f io.ReadSeeker, search []byte, startCursorPos int64) ([]int, error) {
	var indices []int
	var prevFoundIndex int = 0

	cursorPos := startCursorPos

	for prevFoundIndex != -1 {
		index, error := findByteStringInReadSeeker(f, search, cursorPos)

		if error != nil {
			return indices, error
		}

		if index != -1 {
			indexFromStartCursorPos := int(cursorPos-startCursorPos) + int(index)
			indices = append(indices, indexFromStartCursorPos)

			cursorPos += int64(index) + 1

			f.Seek(cursorPos, io.SeekStart)
		}

		prevFoundIndex = index
	}

	// Reset cursor
	f.Seek(startCursorPos, io.SeekStart)

	return indices, nil
}

func findByteStringInReadSeeker(f io.ReadSeeker, search []byte, startCursorPos int64) (int, error) {
	data, err := ioutil.ReadAll(f)

	if err != nil {
		return 0, err
	}

	index := bytes.Index(data, search)

	// Reset cursor
	f.Seek(startCursorPos, io.SeekStart)

	return index, nil
}

func main() {
	fs := http.FileServer(http.Dir("static/"))
	http.Handle("/", http.StripPrefix("/", fs))

	http.HandleFunc("/regionslist", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(getRegionsList())
	})

	http.HandleFunc("/chunkdata", func(w http.ResponseWriter, r *http.Request) {
		var per_page int = getIntParamWithFallback(r, "per_page", 10)
		var page int = getIntParamWithFallback(r, "page", 0)

		var from int = page * per_page
		var to int = (page * per_page) + per_page

		var regions = getRegionsList()[from:to]

		jobs := make(chan RegionReturnData)

		results := make(chan []ChunkReturnData)

		wg := new(sync.WaitGroup)

		for w := 1; w <= per_page; w++ {
			wg.Add(1)
			go getChunkData(jobs, results, wg)
		}

		go func() {
			for _, region := range regions {
				jobs <- region
			}
			close(jobs)
		}()

		go func() {
			wg.Wait()
			close(results)
		}()

		// wg.Wait()
		// close(results)

		var chunkData []ChunkReturnData
		for result := range results {
			chunkData = append(chunkData, result...)
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(chunkData)
	})

	http.ListenAndServe(":8181", nil)
}

func getRegionsList() []RegionReturnData {
	var regionDataArray []RegionReturnData

	entries, _ := os.ReadDir(rootDir)

	for _, entry := range entries {
		fileNameParts := strings.Split(entry.Name(), ".")

		posX, _ := strconv.Atoi(fileNameParts[1])
		posZ, _ := strconv.Atoi(fileNameParts[2])

		regionDataArray = append(regionDataArray, RegionReturnData{
			PosX: posX,
			PosZ: posZ,
		})
	}

	// Sort regions by closeness to 0, 0
	sort.Slice(regionDataArray, func(a, b int) bool {
		return math.Abs(float64(regionDataArray[a].PosX))+math.Abs(float64(regionDataArray[a].PosZ)) < math.Abs(float64(regionDataArray[b].PosX))+math.Abs(float64(regionDataArray[b].PosZ))
	})

	return regionDataArray
}

func getStoragePath(subPath string) (string, error) {
	exePath, err := os.Executable()

	if err != nil {
		panic(err)
	}

	exePath = filepath.Dir(exePath)
	subbedPath := fmt.Sprintf("%s/%s", exePath, subPath)

	_, err = os.Stat(subbedPath)

	return subbedPath, err
}

func getCachedChunkData(region RegionReturnData) []ChunkReturnData {
	var cachedChunkData []ChunkReturnData

	filePath, _ := getStoragePath(fmt.Sprintf("cache/region.%d.%d.json", region.PosX, region.PosZ))
	file, error := os.OpenFile(filePath, os.O_RDONLY, os.ModePerm)

	if error == nil {
		json.NewDecoder(file).Decode(&cachedChunkData)
	}

	return cachedChunkData
}

func setCachedChunkData(region RegionReturnData, chunkData []ChunkReturnData) {
	cachePath, error := getStoragePath("cache")

	if os.IsNotExist(error) {
		os.Mkdir(cachePath, os.ModePerm)
	}

	filePath, _ := getStoragePath(fmt.Sprintf("cache/region.%d.%d.json", region.PosX, region.PosZ))

	file, _ := os.OpenFile(filePath, os.O_CREATE, os.ModePerm)
	defer file.Close()

	file.Truncate(0)
	json.NewEncoder(file).Encode(chunkData)
}

func getChunkData(regions <-chan RegionReturnData, results chan<- []ChunkReturnData, wg *sync.WaitGroup) {
	// Decreasing internal counter for wait-group as soon as goroutine finishes
	defer wg.Done()

	var chunkDataArray []ChunkReturnData

	// start := time.Now()

	for region := range regions {
		chunkDataForRegion := getCachedChunkData(region)

		if len(chunkDataForRegion) > 0 {
			chunkDataArray = append(chunkDataArray, chunkDataForRegion...)
			continue
		}

		regionFileData, error := os.ReadFile(fmt.Sprintf("%s\\r.%d.%d.mca", rootDir, region.PosX, region.PosZ))

		if error != nil {
			continue
		}

		for chunkIndex := range 1024 {
			// The first header block contains the locations of chunk data in the file. Each chunk location is expressed by 3 bytes, and 1 sector count byte.
			var chunkDataOffsetInSectors uint32 = int24BinaryToInt32(regionFileData[chunkIndex*4 : chunkIndex*4+3])

			// This chunk does not exist
			if chunkDataOffsetInSectors == 0 {
				continue
			}

			// The chunk data starts at the chunk location (in 4KiB sectors). The first four bytes contain the data length of this chunk (in bytes).
			chunkDataStart := int64(chunkDataOffsetInSectors * 1024 * 4)
			var chunkDataLengthInBytes uint32 = binary.BigEndian.Uint32(regionFileData[chunkDataStart : chunkDataStart+4])

			if chunkDataLengthInBytes == 0 {
				continue
			}

			// Read the chunk data and stream it using a bytes Reader. Skip over the first 4 bytes of the chunk data (details the chunk data length) and the 5th bit (detail compression type).
			compressedChunkDataBytesBuffer := bytes.NewReader(regionFileData[chunkDataStart+4+1 : chunkDataStart+4+1+int64(chunkDataLengthInBytes)])

			// Uncompress the chunk data (assuming zlib)
			reader, error := zlib.NewReader(compressedChunkDataBytesBuffer)

			// If the chunk data could not be uncompressed (unlikely), skip this chunk.
			if reader == nil || error != nil {
				fmt.Printf("zlib failed")
				continue
			}

			// Read the uncompressed chunk data
			chunkData, _ := ioutil.ReadAll(reader)
			reader.Close()

			// Parse the uncompressed chunk data using an NBT package
			var chunk map[string]interface{}
			error = nbt.Unmarshal(chunkData, &chunk)

			// Get the x,z locations of this chunk
			var xPos = chunk["xPos"]
			var zPos = chunk["zPos"]

			// Get the sections of this chunk
			var sections = chunk["sections"].([]interface{})

			paletteCounts := make(map[string]int)

			// Iterate over the sections of this chunk
			for _, section := range sections {
				// Get the block_states of this chunk
				var blockstates = section.(map[string]interface{})["block_states"]

				if blockstates != nil {
					// Info from https://minecraft.fandom.com/wiki/Chunk_format:
					// - Packed array of 4096 indices, stored in an array of 64-bit ints.
					// - If only one block state is present in the palette, this field is not required and the block fills the whole section.
					// - All indices are the same length: the minimum amount of bits required to represent the largest index in the palette.
					// - These indices have a minimum size of 4 bits.
					// - Since 1.16, the indices are not packed across multiple elements of the array, meaning that if there is no more space in
					//   a given 64-bit integer for the next index, it starts instead at the first (lowest) bit of the next 64-bit element.

					// Info from https://wiki.vg/Chunk_Format#Paletted_Container_structure:
					// - Chunk sections
					//   - Chunk sections are 16x16x16 collections of blocks.
					//   - Chunk sections store blocks, biomes and light data (both block light and sky light).
					//   - Chunk sections can be associated with at most two palettes — one for blocks, one for biomes.
					//   - Chunk sections can contain at maximum 4096 (16×16×16, or 212) unique block state IDs, and 64 (4×4×4) unique biome IDs (highly unlikely).
					// - Registries
					//   - The registries are the primary, protocol-wide mappings from block states and biomes to numeric identifiers.
					//   - The block state registry is hardcoded into Minecraft.
					//   - One block state ID is allocated for each unique block state of a block
					//   - If a block has multiple properties then the number of allocated states is the product of the number of values for each property.
					//   - The block state IDs belonging to a given block are always consecutive. Other than that, the ordering of block states is hardcoded, and somewhat arbitrary.
					//   - The Data Generators system can be used to generate a list of all block state IDs.
					//   - The biome registry is defined at runtime as part of the Registry Data packet sent by the server during the Configuration phase.
					//   - The Notchian server pulls these biome definitions from data packs.
					// - Palettes
					//   - A palette maps a smaller set of IDs within a chunk section to registry IDs
					//   - For example:
					//     - Encoding any of the IDs in the block state registry as of vanilla 1.20.2 requires 15 bits.
					//     - Given that most sections contain only a few different blocks, using 15 bits per block to represent a chunk section that is
					//       only stone, gravel, and air would be extremely wasteful. Instead, a list of registry IDs is sent (for instance, 40 57 0),
					//       and indices into that list — the palette — are sent as the block state or biome values within the chunk (so 40 would be sent
					//       as 0, 57 as 1, and 0 as 2)
					//     - The number of bits used to encode palette indices varies based on the number of indices, and the registry in question.
					//     - If a threshold on the number of unique IDs in the section is exceeded, a palette is not used, and registry IDs are used directly instead.
					// - Heightmaps
					//   - Minecraft uses heightmaps to optimize various operations on both the server and the client.
					//   - All heightmaps share the basic structure of encoding the position of the highest "occupied" block in each column of blocks within a chunk
					//     column. The differences have to do with which blocks are considered to be "occupied".
					//   - Rather than calculating them from the chunk data, the client receives the initial heightmaps it needs from the server. This trades an
					//     increase in network usage for a decrease in client-side processing. Once a chunk is loaded, the client updates its heightmaps based on
					//     block changes independently from the server.

					// var palette = blockstates.(map[string]interface{})["palette"]
					// var blockStatesData = blockstates.(map[string]interface{})["data"]

					// if xPos.(int32) == 0 && zPos.(int32) == 0 {
					// 	if blockStatesData != nil {
					// 		// BlockPos := y*16*16 + z*16 + x;
					// 		// compound Block = Palette[change_array_element_size(BlockStates,Log2(length(Palette)))[BlockPos]];
					// 		// string BlockName = Block.Name;

					// 		// bits per value = max(int(numpy.amax(b)).bit_length(), 1)
					// 		// var largestNumberInBlockStatesData int64

					// 		// for _, blockState := range blockStatesData.([]int64) {
					// 		// 	if blockState > largestNumberInBlockStatesData {
					// 		// 		largestNumberInBlockStatesData = blockState
					// 		// 	}
					// 		// }

					// 		fmt.Println(math.Log2(float64(len(palette.([]interface{})))))
					// 		fmt.Println(palette)
					// 		// fmt.Println(blockStatesData.([]int64)[offset])
					// 	}
					// }
				}

				// Get the biomes of this chunk
				var biomes = section.(map[string]interface{})["biomes"]

				if biomes == nil {
					continue
				}

				// Get the palette of this biome
				var palettes = biomes.(map[string]interface{})["palette"]

				if palettes == nil {
					continue
				}

				// Iterate over the palette of this biome
				for _, palette := range palettes.([]interface{}) {
					// Log the occurence of the biome string in the palette
					paletteCounts[palette.(string)] += 1
				}
			}

			var mostOftOcurringBiome string

			// Determine the most dominant biome in this chunk
			for biome, ocurrences := range paletteCounts {
				if mostOftOcurringBiome == "" {
					mostOftOcurringBiome = biome
				} else if ocurrences > paletteCounts[mostOftOcurringBiome] {
					mostOftOcurringBiome = biome
				}
			}

			chunkDataForRegion = append(chunkDataForRegion, ChunkReturnData{
				PosX:  int(xPos.(int32)),
				PosZ:  int(zPos.(int32)),
				Biome: mostOftOcurringBiome,
			})
		}

		setCachedChunkData(region, chunkDataForRegion)
		chunkDataArray = append(chunkDataArray, chunkDataForRegion...)
	}

	// fmt.Printf("Total Time: %s\n", time.Since(start))

	results <- chunkDataArray
}
