VOID_BLACK = (0, 0, 0)
PLAINS_GREEN = (31, 40, 22)
SNOW_WHITE = (90, 93, 93)
SAND_YELLOW = (88, 82, 64)
BADLANDS_ORANGE = (80, 42, 15)
STONE_GREY = (43, 43, 43)
WATER_BLUE = (27, 31, 73)
NETHER_RED = (51, 26, 26)
SUNFLOWER_YELLOW = (252, 224, 64)
SWAMP_GREEN = (56, 73, 65)
FOREST_GREEN = (49, 51, 27)
CHERRY_PINK = (219, 172, 193)

class BiomeColors(object):
    the_void = VOID_BLACK
    plains = PLAINS_GREEN
    sunflower_plains = SUNFLOWER_YELLOW
    snowy_plains = SNOW_WHITE
    ice_spikes = SNOW_WHITE
    desert = SAND_YELLOW
    swamp = SWAMP_GREEN
    mangrove_swamp = SWAMP_GREEN
    forest = FOREST_GREEN
    flower_forest = FOREST_GREEN
    birch_forest = FOREST_GREEN
    dark_forest = FOREST_GREEN
    old_growth_birch_forest = FOREST_GREEN
    old_growth_pine_taiga = FOREST_GREEN
    old_growth_spruce_taiga = FOREST_GREEN
    taiga = FOREST_GREEN
    snowy_taiga = SNOW_WHITE
    savanna = SAND_YELLOW
    savanna_plateau = SAND_YELLOW
    windswept_hills = PLAINS_GREEN
    windswept_gravelly_hills = PLAINS_GREEN
    windswept_forest = PLAINS_GREEN
    windswept_savanna = SAND_YELLOW
    jungle = PLAINS_GREEN
    sparse_jungle = PLAINS_GREEN
    bamboo_jungle = PLAINS_GREEN
    badlands = BADLANDS_ORANGE
    eroded_badlands = BADLANDS_ORANGE
    wooded_badlands = BADLANDS_ORANGE
    meadow = PLAINS_GREEN
    cherry_grove = CHERRY_PINK
    grove = SNOW_WHITE
    snowy_slopes = SNOW_WHITE
    frozen_peaks = SNOW_WHITE
    jagged_peaks = SNOW_WHITE
    stony_peaks = STONE_GREY
    river = WATER_BLUE
    frozen_river = SNOW_WHITE
    beach = SAND_YELLOW
    snowy_beach = SNOW_WHITE
    stony_shore = STONE_GREY
    warm_ocean = WATER_BLUE
    lukewarm_ocean = WATER_BLUE
    deep_lukewarm_ocean = WATER_BLUE
    ocean = WATER_BLUE
    deep_ocean = WATER_BLUE
    cold_ocean = WATER_BLUE
    deep_cold_ocean = WATER_BLUE
    frozen_ocean = SNOW_WHITE
    deep_frozen_ocean = SNOW_WHITE
    mushroom_fields = PLAINS_GREEN
    dripstone_caves = STONE_GREY
    lush_caves = STONE_GREY
    deep_dark = VOID_BLACK
    nether_wastes = NETHER_RED
    warped_forest = NETHER_RED
    crimson_forest = NETHER_RED
    soul_sand_valley = STONE_GREY
    basalt_deltas = STONE_GREY
    the_end = SAND_YELLOW
    end_highlands = SAND_YELLOW
    end_midlands = SAND_YELLOW
    small_end_islands = SAND_YELLOW
    end_barrens = SAND_YELLOW

    def getColor(self, biomeResourceLocation):
        # Fall back on pure RED
        if biomeResourceLocation is None:
            return (255, 0, 0)

        location = biomeResourceLocation.replace('minecraft:', '')
        return getattr(self, location)
