package main

import (
	"bytes"
	"compress/zlib"
	"embed"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/fs"
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

func int24BinaryToInt32(bytes []byte) uint32 {
	return binary.BigEndian.Uint32(append([]byte{0x00}, bytes...))
}

func getIntParamWithFallback(r *http.Request, paramName string, fallback int) int {
	urlParam := r.URL.Query().Get(paramName)

	if len(urlParam) > 0 {
		// parsedInt, error := strconv.Atoi(urlParam)
		parsedInt, error := strconv.Atoi(urlParam)

		if error == nil {
			return parsedInt
		} else {
			fmt.Println(error)
		}
	}

	return fallback
}

func main() {
	loadConfig()
	readAllRegionFiles()

	staticContent, _ := fs.Sub(staticAssets, "static")
	http.Handle("/", http.FileServer(http.FS(staticContent)))

	http.HandleFunc("/regionslist", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(getRegionsList())
	})

	http.HandleFunc("/palette", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(getPalette())
	})

	http.HandleFunc("/blockdata", func(w http.ResponseWriter, r *http.Request) {
		region_x := getIntParamWithFallback(r, "region_x", 0)
		region_z := getIntParamWithFallback(r, "region_z", 0)

		blockData, error := getCachedBlockData(Region{PosX: region_x, PosZ: region_z})

		w.Header().Set("Content-Type", "application/json")

		if error == nil {
			json.NewEncoder(w).Encode(blockData)
		} else {
			json.NewEncoder(w).Encode(make([]int, 0))
		}
	})

	http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", config.WebserverPort), nil)
}

func readAllRegionFiles() {
	wg := new(sync.WaitGroup)
	jobs := make(chan Region)

	wg.Add(config.BackgroundWorkersCount)

	// Add workers
	for w := 1; w <= config.BackgroundWorkersCount; w++ {
		go getChunkData(jobs, wg)
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

	blockTypesPath, _ := getStoragePath("static/blocks")
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

	file.Truncate(0)
	json.NewEncoder(file).Encode(blockData)
}

func convertToBitString(bytes []byte) string {
	bitString := ""

	for i := 0; i < len(bytes); i++ {
		for j := 0; j < 8; j++ {
			zeroOrOne := bytes[i] >> (7 - j) & 1
			bitString += fmt.Sprintf("%c", '0'+zeroOrOne)
		}
	}

	return bitString
}

func getChunkData(regions <-chan Region, wg *sync.WaitGroup) {
	// Decreasing internal counter for wait-group as soon as goroutine finishes
	defer wg.Done()

	internalPalette := getInversedPalette()

	for region := range regions {
		blockDataForRegion, error := getCachedBlockData(region)

		if error == nil {
			continue
		}

		fmt.Printf("Cache MISS for region[x,z]: [%d,%d]\n", region.PosX, region.PosZ)

		regionFileData, error := os.ReadFile(fmt.Sprintf("%s/region/r.%d.%d.mca", config.PathToWorld, region.PosX, region.PosZ))

		if error != nil {
			fmt.Println(error)
			continue
		}

		for chunkIndex := range 1024 {
			fmt.Println(chunkIndex)

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
			var chunkX = chunk["xPos"].(int32)
			var chunkZ = chunk["zPos"].(int32)

			// Get the sections of this chunk
			var sections = chunk["sections"].([]interface{})

			// paletteCounts := make(map[string]int)

			// Iterate over the sections of this chunk
			for i := len(sections) - 1; i > 0; i-- {
				// Get the block_states of this chunk
				var blockstates = sections[i].(map[string]interface{})["block_states"]

				if blockstates == nil {
					continue
				}

				blockData := blockstates.(map[string]interface{})["data"]
				blockPalette := blockstates.(map[string]interface{})["palette"]
				indexSizeInBits := 4

				if blockData == nil {
					continue
				}

				for x := range 16 {
					for z := range 16 {
						var chunkXInRegion int32
						var chunkZInRegion int32

						if chunkX < 0 {
							chunkXInRegion = 31 - (int32(math.Abs(float64(chunkX))) % 32)
						} else {
							chunkXInRegion = chunkX % 32
						}

						if chunkZ < 0 {
							chunkZInRegion = 31 - (int32(math.Abs(float64(chunkZ))) % 32)
						} else {
							chunkZInRegion = chunkZ % 32
						}

						blockXInRegion := int(chunkXInRegion*16) + x
						blockZInRegion := int(chunkZInRegion*16) + z

						// a block (with presumable a higher y-value) was already loaded for these x,z coords.
						if blockDataForRegion[(blockZInRegion*32*16)+blockXInRegion] != 0 {
							continue
						}

						for y := range 16 {
							blockInfoForCoordsBitIndexGlobal := ((y * 16 * 16) + (z * 16) + (x)) * indexSizeInBits
							targetInt64Index := blockInfoForCoordsBitIndexGlobal / 64
							blockInfoForCoordsBitIndexLocal := blockInfoForCoordsBitIndexGlobal - (targetInt64Index * 64)

							targetInt64 := uint64(blockData.([]int64)[targetInt64Index])

							b := make([]byte, 8)
							binary.BigEndian.PutUint64(b, targetInt64)
							bitString := convertToBitString(b)

							myBlockPalette := bitString[64-blockInfoForCoordsBitIndexLocal-indexSizeInBits : 64-blockInfoForCoordsBitIndexLocal]

							myPaletteIndex, _ := strconv.ParseInt(myBlockPalette, 2, 0)
							blockType := blockPalette.([]interface{})[myPaletteIndex].(map[string]interface{})["Name"]

							if blockType.(string) != "minecraft:air" {
								trimmedBlockType := blockType.(string)[10:len(blockType.(string))]

								blockDataForRegion[(blockZInRegion*32*16)+blockXInRegion] = internalPalette[trimmedBlockType]
							}
						}
					}
				}
			}
		}

		setCachedBlockData(region, blockDataForRegion)
	}
}
