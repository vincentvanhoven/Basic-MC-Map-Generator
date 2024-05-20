# Basic-MC-Map-Generator

A small Python program that reads the region files of a Minecraft world and renders a very basic map of it.
This software was specifically written for rendering rather large worlds in a small timeframe.

In a personal benchmark using 50 processes, it rendered a 250+ GB world in about 10 minutes.

## Usage

```python3 main.py {path_to_region_files} {amount_of_processes} {ignore_cache}```

Where `{path_to_region_files}` is the absolute path to the directory containing the region files of your Minecraft world, `{amount_of_processes}` is the amount of processes that the program will start to read out the region files, and `{ignore_cache}` can be used to skip cached chunk data and re-read the region files (`False`, by default).

A note about `{amount_or_processes}`: a higher number is not always faster, and different computers will have different sweet spots.
I recommend trying out different numbers to see which yields the fastest results.

## Why?

With the number of different software options for Minecraft map generation, you may wonder what the point of this little program this.
It was created to have very few features, but to be rather fast instead.
I have a Minecraft world that is only going to grow, and every now and then I want to see a rough preview of what the whole thing looks like.
