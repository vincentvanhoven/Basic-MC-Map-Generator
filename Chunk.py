import sys

class Chunk:
    def __init__(self, dataOffset, timestamp = None, isLoaded = False, x = 0, z =0):
        self.dataOffset = dataOffset
        self.timestamp = timestamp
        self.isLoaded = isLoaded
        self.x = x
        self.z = z
