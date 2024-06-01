package main

type Chunk struct {
	X           int
	Z           int
	BiomeNumber string
}

type Region struct {
	PosX int
	PosZ int
}

type Block struct {
	PosX      int
	PosZ      int
	BlockType string
}

var CustomBiomeColors = map[string][]int{
	"VOID_BLACK":       {0, 0, 0},
	"PLAINS_GREEN":     {31, 40, 22},
	"SNOW_WHITE":       {90, 93, 93},
	"SAND_YELLOW":      {88, 82, 64},
	"BADLANDS_ORANGE":  {80, 42, 15},
	"STONE_GREY":       {43, 43, 43},
	"WATER_BLUE":       {27, 31, 73},
	"NETHER_RED":       {51, 26, 26},
	"SUNFLOWER_YELLOW": {252, 224, 64},
	"SWAMP_GREEN":      {56, 73, 65},
	"FOREST_GREEN":     {49, 51, 27},
	"CHERRY_PINK":      {219, 172, 193},
}

var BiomeColors = map[string]struct {
	Biome string
	Color []int
}{
	// The values below were taken from https://github.com/toolbox4minecraft/amidst/wiki/Biome-Color-Table (2024-05-20)
	"0":  {Biome: "minecraft:ocean", Color: []int{0, 0, 112}},
	"1":  {Biome: "minecraft:plains", Color: []int{141, 179, 96}},
	"2":  {Biome: "minecraft:desert", Color: []int{250, 148, 24}},
	"3":  {Biome: "minecraft:mountains", Color: []int{96, 96, 96}},
	"4":  {Biome: "minecraft:forest", Color: []int{5, 102, 33}},
	"5":  {Biome: "minecraft:taiga", Color: []int{11, 102, 89}},
	"6":  {Biome: "minecraft:swamp", Color: []int{7, 249, 178}},
	"7":  {Biome: "minecraft:river", Color: []int{0, 0, 255}},
	"8":  {Biome: "minecraft:nether_wastes", Color: []int{191, 59, 59}},
	"9":  {Biome: "minecraft:the_end", Color: []int{128, 128, 255}},
	"10": {Biome: "minecraft:frozen_ocean", Color: []int{112, 112, 214}},
	"11": {Biome: "minecraft:frozen_river", Color: []int{160, 160, 255}},
	"12": {Biome: "minecraft:snowy_tundra", Color: []int{255, 255, 255}},
	"13": {Biome: "minecraft:snowy_mountains", Color: []int{160, 160, 160}},
	"14": {Biome: "minecraft:mushroom_fields", Color: []int{255, 0, 255}},
	"15": {Biome: "minecraft:mushroom_field_shore", Color: []int{160, 0, 255}},
	"16": {Biome: "minecraft:beach", Color: []int{250, 222, 85}},
	"17": {Biome: "minecraft:desert_hills", Color: []int{210, 95, 18}},
	"18": {Biome: "minecraft:wooded_hills", Color: []int{34, 85, 28}},
	"19": {Biome: "minecraft:taiga_hills", Color: []int{22, 57, 51}},
	"20": {Biome: "minecraft:mountain_edge", Color: []int{114, 120, 154}},
	"21": {Biome: "minecraft:jungle", Color: []int{83, 123, 9}},
	"22": {Biome: "minecraft:jungle_hills", Color: []int{44, 66, 5}},
	"23": {Biome: "minecraft:jungle_edge", Color: []int{98, 139, 23}},
	"24": {Biome: "minecraft:deep_ocean", Color: []int{0, 0, 48}},
	"25": {Biome: "minecraft:stone_shore", Color: []int{162, 162, 132}},
	"26": {Biome: "minecraft:snowy_beach", Color: []int{250, 240, 192}},
	"27": {Biome: "minecraft:birch_forest", Color: []int{48, 116, 68}},
	"28": {Biome: "minecraft:birch_forest_hills", Color: []int{31, 95, 50}},
	"29": {Biome: "minecraft:dark_forest", Color: []int{64, 81, 26}},
	"30": {Biome: "minecraft:snowy_taiga", Color: []int{49, 85, 74}},
	"31": {Biome: "minecraft:snowy_taiga_hills", Color: []int{36, 63, 54}},
	"32": {Biome: "minecraft:giant_tree_taiga", Color: []int{89, 102, 81}},
	"33": {Biome: "minecraft:giant_tree_taiga_hills", Color: []int{9, 79, 62}},
	"34": {Biome: "minecraft:wooded_mountains", Color: []int{80, 112, 80}},
	"35": {Biome: "minecraft:savanna", Color: []int{189, 178, 95}},
	"36": {Biome: "minecraft:savanna_plateau", Color: []int{167, 157, 100}},
	"37": {Biome: "minecraft:badlands", Color: []int{217, 69, 21}},
	"38": {Biome: "minecraft:wooded_badlands_plateau", Color: []int{6, 151, 101}},
	"39": {Biome: "minecraft:badlands_plateau", Color: []int{202, 140, 101}},
	"40": {Biome: "minecraft:small_end_islands", Color: []int{128, 128, 255}},
	"41": {Biome: "minecraft:end_midlands", Color: []int{128, 128, 255}},
	"42": {Biome: "minecraft:end_highlands", Color: []int{128, 128, 255}},
	"43": {Biome: "minecraft:end_barrens", Color: []int{128, 128, 255}},
	"44": {Biome: "minecraft:warm_ocean", Color: []int{0, 0, 172}},
	"45": {Biome: "minecraft:lukewarm_ocean", Color: []int{0, 0, 144}},
	"46": {Biome: "minecraft:cold_ocean", Color: []int{32, 32, 112}},
	"47": {Biome: "minecraft:deep_warm_ocean", Color: []int{0, 0, 80}},
	"48": {Biome: "minecraft:deep_lukewarm_ocean", Color: []int{0, 0, 64}},
	"49": {Biome: "minecraft:deep_cold_ocean", Color: []int{32, 32, 56}},
	"50": {Biome: "minecraft:deep_frozen_ocean", Color: []int{64, 64, 144}},
	"51": {Biome: "minecraft:the_void", Color: []int{0, 0, 0}},
	"52": {Biome: "minecraft:sunflower_plains", Color: []int{181, 219, 136}},
	"53": {Biome: "minecraft:desert_lakes", Color: []int{255, 188, 64}},
	"54": {Biome: "minecraft:gravelly_mountains", Color: []int{136, 136, 136}},
	"55": {Biome: "minecraft:flower_forest", Color: []int{45, 142, 73}},
	"56": {Biome: "minecraft:taiga_mountains", Color: []int{51, 142, 129}},
	"57": {Biome: "minecraft:swamp_hills", Color: []int{47, 255, 218}},
	"58": {Biome: "minecraft:ice_spikes", Color: []int{180, 220, 220}},
	"59": {Biome: "minecraft:modified_jungle", Color: []int{123, 163, 49}},
	"60": {Biome: "minecraft:modified_jungle_edge", Color: []int{138, 179, 63}},
	"61": {Biome: "minecraft:tall_birch_forest", Color: []int{88, 156, 108}},
	"62": {Biome: "minecraft:tall_birch_hills", Color: []int{71, 135, 90}},
	"63": {Biome: "minecraft:dark_forest_hills", Color: []int{104, 121, 66}},
	"64": {Biome: "minecraft:snowy_taiga_mountains", Color: []int{89, 125, 114}},
	"65": {Biome: "minecraft:giant_spruce_taiga", Color: []int{129, 142, 121}},
	"66": {Biome: "minecraft:giant_spruce_taiga_hills", Color: []int{109, 119, 102}},
	"67": {Biome: "minecraft:shattered_savanna", Color: []int{229, 218, 135}},
	"68": {Biome: "minecraft:shattered_savanna_plateau", Color: []int{207, 197, 140}},
	"69": {Biome: "minecraft:eroded_badlands", Color: []int{255, 109, 61}},
	"70": {Biome: "minecraft:modified_wooded_badlands_plateau", Color: []int{216, 191, 141}},
	"71": {Biome: "minecraft:modified_badlands_plateau", Color: []int{242, 180, 141}},
	"72": {Biome: "minecraft:bamboo_jungle", Color: []int{118, 142, 20}},
	"73": {Biome: "minecraft:bamboo_jungle_hills", Color: []int{59, 71, 10}},
	"74": {Biome: "minecraft:soul_sand_valley", Color: []int{94, 56, 48}},
	"75": {Biome: "minecraft:crimson_forest", Color: []int{221, 8, 8}},
	"76": {Biome: "minecraft:warped_forest", Color: []int{73, 144, 123}},
	"77": {Biome: "minecraft:basalt_deltas", Color: []int{64, 54, 54}},

	// The values below are custom
	"78": {Biome: "minecraft:snowy_plains", Color: CustomBiomeColors["SNOW_WHITE"]},
	"79": {Biome: "minecraft:mangrove_swamp", Color: CustomBiomeColors["SWAMP_GREEN"]},
	"80": {Biome: "minecraft:old_growth_birch_forest", Color: CustomBiomeColors["FOREST_GREEN"]},
	"81": {Biome: "minecraft:old_growth_pine_taiga", Color: CustomBiomeColors["FOREST_GREEN"]},
	"82": {Biome: "minecraft:old_growth_spruce_taiga", Color: CustomBiomeColors["FOREST_GREEN"]},
	"83": {Biome: "minecraft:windswept_hills", Color: CustomBiomeColors["PLAINS_GREEN"]},
	"84": {Biome: "minecraft:windswept_gravelly_hills", Color: CustomBiomeColors["PLAINS_GREEN"]},
	"85": {Biome: "minecraft:windswept_forest", Color: CustomBiomeColors["PLAINS_GREEN"]},
	"86": {Biome: "minecraft:windswept_savanna", Color: CustomBiomeColors["SAND_YELLOW"]},
	"87": {Biome: "minecraft:sparse_jungle", Color: CustomBiomeColors["PLAINS_GREEN"]},
	"88": {Biome: "minecraft:wooded_badlands", Color: CustomBiomeColors["BADLANDS_ORANGE"]},
	"89": {Biome: "minecraft:meadow", Color: CustomBiomeColors["PLAINS_GREEN"]},
	"90": {Biome: "minecraft:cherry_grove", Color: CustomBiomeColors["CHERRY_PINK"]},
	"91": {Biome: "minecraft:grove", Color: CustomBiomeColors["SNOW_WHITE"]},
	"92": {Biome: "minecraft:snowy_slopes", Color: CustomBiomeColors["SNOW_WHITE"]},
	"93": {Biome: "minecraft:frozen_peaks", Color: CustomBiomeColors["SNOW_WHITE"]},
	"94": {Biome: "minecraft:jagged_peaks", Color: CustomBiomeColors["SNOW_WHITE"]},
	"95": {Biome: "minecraft:stony_peaks", Color: CustomBiomeColors["STONE_GREY"]},
	"96": {Biome: "minecraft:stony_shore", Color: CustomBiomeColors["STONE_GREY"]},
	"97": {Biome: "minecraft:dripstone_caves", Color: CustomBiomeColors["STONE_GREY"]},
	"98": {Biome: "minecraft:lush_caves", Color: CustomBiomeColors["STONE_GREY"]},
	"99": {Biome: "minecraft:deep_dark", Color: CustomBiomeColors["VOID_BLACK"]},
}
