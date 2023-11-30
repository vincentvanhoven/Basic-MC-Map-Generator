from PIL import Image, ImageDraw
from RegionFilesReader import RegionFilesReader
import datetime
import sys
import os
import time

from TerminalColors import TerminalColors


if __name__ == '__main__':
    dateTimeStart = datetime.datetime.now()

    sourcePath = sys.argv[1]
    ignoreCache = False
    if len(sys.argv) > 2:
        ignoreCache = sys.argv[2] if sys.argv[2] else False

    outputDir = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'output')

    if os.path.exists(outputDir) == False:
        os.mkdir(outputDir)

    outputFilePath = os.path.join(outputDir, f'map_{time.time()}.png')

    regionFilesReader = RegionFilesReader()
    regionFilesReader.iterateDirectory(sourcePath, ignoreCache)

    print(f'Read {len(regionFilesReader.chunks)} chunks from the region files or cache. Now generating the render...')

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

    print(f'Chunk range runs from [{minX}, {minZ}] to [{maxX}, {maxZ}]')

    tileSize = 12
    imageWidth = (abs(minX) + maxX) * tileSize
    imageHeight = (abs(minZ) + maxZ) * tileSize

    im = Image.new(mode="RGB", size=(imageWidth, imageHeight))

    draw = ImageDraw.Draw(im)

    for chunk in regionFilesReader.chunks:
        color = (254, 0, 0)

        draw.rectangle(
            (
                (imageWidth / 2) + (chunk.x * tileSize),
                # The height position has to be offset by (- tileSize * 2) to render properly, for some reason. Unsure why.
                (imageHeight / 2) + (chunk.z * tileSize) - (tileSize * 2),
                (imageWidth / 2) + (chunk.x * tileSize) + tileSize,
                (imageHeight / 2) + (chunk.z * tileSize) - tileSize,
            ),
            fill=color
        )

    im.save(outputFilePath)

    print(f'{TerminalColors.OKGREEN}The file was generated in: {(datetime.datetime.now() - dateTimeStart).total_seconds()} seconds and can be found here:{TerminalColors.ENDC}')
    print(outputFilePath)

