class Chunk:
    def __init__(self, dataOffset, biome = None, timestamp = None, isLoaded = False, x = 0, z = 0):
        self.dataOffset = dataOffset
        self.timestamp = timestamp
        self.isLoaded = isLoaded
        self.x = x
        self.z = z
        self.biome = biome
