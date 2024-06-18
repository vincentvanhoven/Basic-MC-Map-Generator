package main

import (
	"bytes"
	"compress/zlib"
	"embed"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	_ "image/png"

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

//go:embed static/*
var staticAssets embed.FS

var config Config
var imageCache map[string]image.Image = make(map[string]image.Image)

func getIntPathValue(w http.ResponseWriter, r *http.Request, paramName string) (int, error) {
	urlParam := r.PathValue(paramName)

	if len(urlParam) > 0 {
		urlPathParam, error := strconv.Atoi(urlParam)

		if error == nil {
			return urlPathParam, error
		}
	}

	// Set content & status headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	// Prepare a JSON message
	responseJson := make(map[string]string)
	responseJson["message"] = fmt.Sprintf("URL path parameter %s must be present and must be numerical.", paramName)

	// Encode the JSON message unto the response
	json.NewEncoder(w).Encode(responseJson)

	return 0, fmt.Errorf("path param %s is missing", paramName)
}

func main() {
	loadConfig()
	readAllRegionFiles()
	preloadTileImageCache()

}

func readAllRegionFiles() {
	wg := new(sync.WaitGroup)
	jobs := make(chan Region)

	wg.Add(config.BackgroundWorkersCount)

	// Add workers
	for w := 1; w <= config.BackgroundWorkersCount; w++ {
		go getRegionData(jobs, wg)
	}

	// Queue jobs
	go func() {
		for _, region := range getRegionsList() {
			jobs <- region
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
	}()
}

func loadConfig() {
	var defaultConfig Config = Config{
		PathToWorld:            "c:/users/vincent/desktop/mc server/anarchy/world",
		WebserverPort:          8181,
		BackgroundWorkersCount: 10,
	}

	filePath, error := getStoragePath("config.json")

	if os.IsNotExist(error) {
		config = defaultConfig
		configFile, _ := os.Create(filePath)

		jsonParser := json.NewEncoder(configFile)
		jsonParser.SetIndent("", "  ")
		jsonParser.Encode(config)

		fmt.Println("Default config file loaded.")
		return
	}

	configFile, error := os.OpenFile(filePath, os.O_RDONLY, os.ModePerm)

	if error != nil {
		panic(error)
	}

	defer configFile.Close()

	json.NewDecoder(configFile).Decode(&config)

	fmt.Println("Config file loaded.")
}

func getPalette() map[int]string {
	var regionDataArray map[int]string = make(map[int]string)

	blockTypesPath, _ := getStoragePath("static/resourcepack/textures/block")
	entries, _ := os.ReadDir(blockTypesPath)

	for index, entry := range entries {
		fileNameParts := strings.Split(entry.Name(), ".")

		if fileNameParts[len(fileNameParts)-1] != "png" {
			continue
		}

		regionDataArray[index+1] = fileNameParts[0]
	}

	return regionDataArray
}

func getInversedPalette() map[string]int {
	inversedPalette := make(map[string]int)

	for index, value := range getPalette() {
		inversedPalette[value] = index
	}

	return inversedPalette
}

func getRegionsList() []Region {
	var regionDataArray []Region

	entries, _ := os.ReadDir(fmt.Sprintf("%s/region", config.PathToWorld))

	for _, entry := range entries {
		fileNameParts := strings.Split(entry.Name(), ".")

		if fileNameParts[len(fileNameParts)-1] != "mca" {
			continue
		}

		posX, _ := strconv.Atoi(fileNameParts[1])
		posZ, _ := strconv.Atoi(fileNameParts[2])

		regionDataArray = append(regionDataArray, Region{
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

func getCachedBlockData(region Region) ([]int, error) {
	var cachedBlockData []int = make([]int, 32*32*16*16)

	filePath, _ := getStoragePath(fmt.Sprintf("cache/region.%d.%d.json", region.PosX, region.PosZ))
	file, error := os.OpenFile(filePath, os.O_RDONLY, os.ModePerm)

	if error == nil {
		fmt.Printf("READ [cache/json] for region [%d, %d]\n", region.PosX, region.PosZ)
		json.NewDecoder(file).Decode(&cachedBlockData)
	}

	return cachedBlockData, error
}

func setCachedBlockData(region Region, blockData []int) {
	cachePath, error := getStoragePath("cache")

	if os.IsNotExist(error) {
		os.Mkdir(cachePath, os.ModePerm)
	}

	filePath, _ := getStoragePath(fmt.Sprintf("cache/region.%d.%d.json", region.PosX, region.PosZ))
	file, error := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, os.ModePerm)

	if error != nil {
		panic(error)
	}

	defer file.Close()

	fmt.Printf("WRITE [cache/json] for region [%d, %d]\n", region.PosX, region.PosZ)

	file.Truncate(0)
	json.NewEncoder(file).Encode(blockData)
}

func getRegionData(regions <-chan Region, wg *sync.WaitGroup) {
	// Decreasing internal counter for wait-group as soon as goroutine finishes
	defer wg.Done()

	internalPalette := getInversedPalette()

	for region := range regions {
		blockDataForRegion, error := getCachedBlockData(region)

		if error == nil {
			continue
		}

		regionFileData, error := os.ReadFile(fmt.Sprintf("%s/region/r.%d.%d.mca", config.PathToWorld, region.PosX, region.PosZ))

		if error != nil {
			fmt.Println(error)
			continue
		}

		for chunkX := range 32 {
			for chunkZ := range 32 {
				var chunkIndex = chunkX + (chunkZ << 5)

				// The first header block contains the locations of chunk data in the file. Each chunk location is expressed by 3 bytes, and 1 sector count byte.
				var chunkDataOffsetInSectors uint32 = binary.BigEndian.Uint32(append([]byte{0x00}, regionFileData[chunkIndex*4:chunkIndex*4+3]...))

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

				getBlockDataForChunk(chunk, chunkX, chunkZ, &blockDataForRegion, internalPalette)
			}
		}

		setCachedBlockData(region, blockDataForRegion)
		writeRegionImage(region, blockDataForRegion)
	}
}

func getBlockDataForChunk(chunk map[string]interface{}, chunkX int, chunkZ int, blockDataForRegion *[]int, internalPalette map[string]int) {
	// Get the lowest Y section position in the chunk
	chunkLowestYSectionPos := chunk["yPos"].(int32)

	// Get the sections of this chunk
	var sections = chunk["sections"].([]interface{})
	var motionBlockingHeightMap = chunk["Heightmaps"].(map[string]interface{})["MOTION_BLOCKING"]

	if motionBlockingHeightMap == nil {
		// This chunk ain't ready
		return
	}

	indexSizeInBits := 4

	for blockX := range 16 {
		for blockZ := range 16 {
			// Calculate the offsets for the heightmap location of the current block X/Z coords
			blockIndex := (blockX) + (blockZ << 4)
			blockIndexInHeightMap := blockIndex / 7
			blockIndexInHeightMapValue := blockIndex % 7

			// Get the packed height map value
			heightMapPackedValue := uint64(motionBlockingHeightMap.([]int64)[blockIndexInHeightMap])
			// Shift to the right by index * 9 (>>), and read the last 9 bits (&)
			blockHeightMapValue := (heightMapPackedValue >> (blockIndexInHeightMapValue * 9)) & uint64(0x1FF)

			// From the wiki: `highestBlockY = (chunk.yPos * 16) - 1 + heightmap_entry_value`
			blockHeightValue := (int(chunkLowestYSectionPos) * 16) - 1 + int(blockHeightMapValue)

			// The height value may indicate that there is no block in this heightmap for these X,Z coords.
			if blockHeightValue < int(chunkLowestYSectionPos) {
				continue
			}

			// Determine the chunk section that contains the block from the heightmap
			blockSectionIndex := int8(blockHeightValue / 16)

			// Determine the local chunk section coordinates for the block
			var blockInTheSection int
			if blockHeightValue >= 0 {
				blockInTheSection = blockHeightValue % 16
			} else {
				blockInTheSection = (16 + (blockHeightValue % 16)) % 16
			}

			// Get the chunk section that contains the block from the heightmap
			var blockSection map[string]interface{}
			for _, section := range sections {
				if section.(map[string]interface{})["Y"].(int8) == blockSectionIndex {
					blockSection = section.(map[string]interface{})
				}
			}

			// Fetch the data for this chunk section
			blockStates := blockSection["block_states"]
			blockData := blockStates.(map[string]interface{})["data"]
			blockPalette := blockStates.(map[string]interface{})["palette"]

			// Get the index of the block in a 'flattened' chunk coordinate system
			blockIndexInData := ((blockInTheSection * 16 * 16) + (blockZ * 16) + (blockX))
			// Divide by the amount of values packed into each uint64 in the blockData array to get the blockData index
			blockIndexInPackedData := blockIndexInData / (64 / indexSizeInBits)
			// Get the remainder to determine the byte index within the uint64 value (from the blockData array) to get the corresponding palette index
			blockSubIndexInPackedData := blockIndexInData % (64 / indexSizeInBits)

			// Get the packed block data, which is an uint64 containing n amount of pallete indices, with n being:
			// - 64 bits / minimum amount of bits to represent the largest palette index.
			// The minimum amount of bits is often 4, resulting in 64/4 = 16 values.
			packedBlockData := uint64(blockData.([]int64)[blockIndexInPackedData])

			// Extract the relevant bits from the uint64 as follows (assuming an index size of 4):
			// - Shift the uint64 by n * 4. This results in the desired bits being at the lowest 4 bit indices. Assuming n = 1:
			//   - Before: 00000000 00110000 00000011 11101111 00001100 10000011 01111111 11110100
			//   - After:  00000000 00000011 00000000 00111110 11110000 11001000 00110111 11111111
			// - Apply the AND operator to extract the lowest 4 bits.
			//   - Before: 00000000 00000011 00000000 00111110 11110000 11001000 00110111 11111111
			//   - After:  00000000 00000000 00000000 00000000 00000000 00000000 00000000 00001111
			unpackedBlockData := (packedBlockData >> (blockSubIndexInPackedData * indexSizeInBits)) & ((1 << indexSizeInBits) - 1)

			// Use the extracted palette index to fetch the block type ('minecraft:id' string)
			blockType := blockPalette.([]interface{})[unpackedBlockData].(map[string]interface{})["Name"]

			// Cut off the 'minecraft:' part of the block type string.
			trimmedBlockType := blockType.(string)[10:len(blockType.(string))]

			// Attempt to get the `_top` block variant from the palette index
			var paletteIndex int = internalPalette[fmt.Sprintf("%s_top", trimmedBlockType)]

			// If the `_top` variant does not exist, use normal one instead
			if paletteIndex == 0 {
				paletteIndex = internalPalette[trimmedBlockType]
			}

			// Determine the local region block coordinates
			blockXInRegion := int(chunkX*16) + blockX
			blockZInRegion := int(chunkZ*16) + blockZ

			// Write the block palette value to the region data
			(*blockDataForRegion)[(blockZInRegion*32*16)+blockXInRegion] = paletteIndex
		}
	}
}

func preloadTileImageCache() {
	for _, blockName := range getPalette() {
		blockTexturePath := fmt.Sprintf("static/resourcepack/textures/block/%s.png", blockName)

		f, err := os.Open(blockTexturePath)

		if err != nil {
			continue
		}

		defer f.Close()
		f.Seek(0, 0)
		blockTexture, _, err := image.Decode(f)

		if err != nil {
			panic(err)
		}

		imageCache[blockName] = blockTexture
	}
}

func writeRegionImage(region Region, regionData []int) {
	palette := getPalette()

	// Rectangle for the full image
	rect := image.Rectangle{image.Point{0, 0}, image.Point{512 * 16, 512 * 16}}
	// The full image
	rgba := image.NewRGBA(rect)

	for z := range 512 {
		for x := range 512 {
			blockPalleteIndex := regionData[(z*32*16)+x]
			blockValue := palette[blockPalleteIndex]
			blockTexture := imageCache[blockValue]

			if blockTexture == nil {
				continue
			}

			// starting position for this tile
			tileStartingPoint := image.Point{blockTexture.Bounds().Dx() * x, blockTexture.Bounds().Dy() * (z)}

			// new rectangle for this tile
			tileRectangle := image.Rectangle{tileStartingPoint, tileStartingPoint.Add(blockTexture.Bounds().Size())}

			draw.Draw(rgba, tileRectangle, blockTexture, image.Point{0, 0}, draw.Src)
		}
	}

	cachePath, error := getStoragePath("cache")

	if os.IsNotExist(error) {
		os.Mkdir(cachePath, os.ModePerm)
	}

	filePath, _ := getStoragePath(fmt.Sprintf("cache/region.%d.%d.jpeg", region.PosX, region.PosZ))
	file, error := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, os.ModePerm)

	if error != nil {
		panic(error)
	}

	defer file.Close()

	fmt.Printf("WRITE [cache/jpeg] for region [%d, %d]\n", region.PosX, region.PosZ)

	file.Truncate(0)
	jpeg.Encode(file, rgba, &jpeg.Options{Quality: 10})
}
