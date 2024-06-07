import './style.css';

import { Chunk } from "./Chunk";
import { Region } from "./Region";
import Vector from "./Vector";
import CacheManager from './CacheManager';

class MapHandler {
    private cache: CacheManager = new CacheManager();
    private renderTileSize: number = 16;
    private regions: Region[] = [];
    private loadedImages: {[key:string]: HTMLImageElement} = {};
    private palette: {[key:number]: string} = [];
    private canvas: HTMLCanvasElement;
    private canvasContext: CanvasRenderingContext2D;
    private imageCreationCanvas: HTMLCanvasElement;
    private imageCreationCanvasContext: CanvasRenderingContext2D;

    private isMouseDown: boolean = false;
    private dragOffsetInPx: Vector = new Vector(0, 0);
    private zoomLevel: number = 1000;

    constructor() {
        this.canvas = <HTMLCanvasElement>document.getElementById("canvas");
        this.canvas.width = window.innerWidth;
        this.canvas.height = window.innerHeight;

        this.canvasContext = <CanvasRenderingContext2D>this.canvas.getContext("2d");
        this.canvasContext.imageSmoothingEnabled = false;

        this.fetchAllRegions();

        // window.setInterval(this.render.bind(this), 1000/60);
        window.setInterval(this.render.bind(this), 1000/1);

        window.addEventListener('resize', this.onResizeEvent.bind(this));
        window.addEventListener('mousedown', this.onMouseDownEvent.bind(this));
        window.addEventListener('mousemove', this.onMouseMoveEvent.bind(this));
        window.addEventListener('mouseup', this.onMouseUpEvent.bind(this));
        document.addEventListener('wheel', this.onWheelEvent.bind(this));

        this.imageCreationCanvas = document.createElement('canvas');
        this.imageCreationCanvas.width = 32;
        this.imageCreationCanvas.height = 32;
        this.imageCreationCanvasContext = <CanvasRenderingContext2D>this.imageCreationCanvas.getContext("2d");
    }

    async fetchAllRegions() {
        await fetch('/regionslist')
            .then((response) => {
                return response.json();
            })
            .then(async (json) => {
                this.regions = json;
            });
        
        await fetch('/palette')
            .then((response) => {
                return response.json()
            })
            .then(async (json) => {
                this.palette = json;
            });
    
        for(let index = 0; index < this.regions.length; index++) {
            await this.fetchBlockData(this.regions[index]);
        }
    }

    async fetchBlockData(region: Region) {
        return fetch(`/blockdata?region_x=${region.PosX}&region_z=${region.PosZ}`)
            .then((response) => {
                return response.json();
            })
            .then((json) => {
                region.blockStates = json;
                return json;
            })
            .catch((err) => {
                throw err;
            });
    }

    onResizeEvent(_event: Event) {
        this.canvas.width = window.innerWidth;
        this.canvas.height = window.innerHeight;
        this.cache.purgeAll();
    }

    onMouseDownEvent(_event: MouseEvent) {
        this.isMouseDown = true;
    }

    onMouseMoveEvent(event: MouseEvent) {
        if(this.isMouseDown) {
            this.dragOffsetInPx.x += event.movementX / (this.zoomLevel / 1000);
            this.dragOffsetInPx.y += event.movementY / (this.zoomLevel / 1000);

            this.cache.purgeAll();
        }
    }

    onMouseUpEvent(_event: MouseEvent) {
        this.isMouseDown = false;
    }

    onWheelEvent(event: WheelEvent) {
        this.zoomLevel -= event.deltaY;
    }

    getCanvasCenterOffset(): Vector {
        return <Vector>this.cache.remember('canvas-center-offset', () => {
            return new Vector(
                this.canvas.width / 2,
                this.canvas.height / 2,
            );
        });
    }

    getChunkRenderSize(): Vector {
        return <Vector>this.cache.remember('chunk-render-size', () => {
            let chunkRenderSize = new Vector(this.renderTileSize, this.renderTileSize);
            return chunkRenderSize.multiply(this.zoomLevel / 1000);
        });
    }

    getTotalChunkOffset(): Vector {
        return <Vector>this.cache.remember('chunk-total-offset', () => {
            return new Vector(
                this.getCanvasCenterOffset().x + this.dragOffsetInPx.x,
                this.getCanvasCenterOffset().y + this.dragOffsetInPx.y,
            );
        });
    }

    getChunkRenderPosition(chunk: Chunk): Vector {
        return <Vector>this.cache.remember(`chunk-render-pos-${chunk.PosX}-${chunk.PosZ}`, () => {
            let chunkPosition = new Vector(chunk.PosX, -chunk.PosZ);

            return chunkPosition
                // Multiply the position by the tilesize
                .multiply(this.renderTileSize)
                // Add the global offset
                .add(this.getTotalChunkOffset())
                .multiply(this.zoomLevel / 1000);
        });
    }

    render() {
        this.canvasContext.save();

        // this.canvasContext.clearRect(0, 0, this.canvas.width, this.canvas.height)
        this.canvasContext.fillStyle = `black`;
        this.canvasContext.fillRect(0, 0, this.canvas.width, this.canvas.height);

        this.canvasContext.scale(this.zoomLevel / 1000, this.zoomLevel / 1000)

        // Determine compensation for keeping negative coords inside viewport.
        let totalOffset: Vector = this.getTotalChunkOffset();

        this.regions.forEach((region: Region) => {
            if(!region.blockStates) {
                return;
            }

            let regionWidthHeight = 32*16;
            let regionRenderSize = regionWidthHeight * this.renderTileSize;
            let regionRenderPos: Vector = new Vector(region.PosX, region.PosZ);
            regionRenderPos.multiply(regionRenderSize).add(totalOffset);

            if(
                regionRenderPos.x + regionRenderSize < 0
                || regionRenderPos.x > this.canvas.clientWidth
                || regionRenderPos.y + regionRenderSize < 0
                || regionRenderPos.y > this.canvas.clientHeight
            ) {
                return;
            }


            region.blockStates.forEach((paletteIndex, blockIndex) => {
                let blockState = this.palette[paletteIndex];

                if(blockState) {
                    if(!this.loadedImages[blockState]) {
                        this.loadedImages[blockState] = new Image();
                        this.loadedImages[blockState].src = `resourcepack/textures/block/${blockState}.png`;
                    } else {
                        let blockRenderPos = new Vector(
                            (blockIndex % regionWidthHeight) * this.renderTileSize,
                            Math.floor(blockIndex / regionWidthHeight) * this.renderTileSize,
                        ).add(regionRenderPos);

                        if(((blockIndex % regionWidthHeight) * this.renderTileSize) % 1 > 0) {
                            console.log((blockIndex % regionWidthHeight) * this.renderTileSize)
                        }

                        this.canvasContext.drawImage(this.loadedImages[blockState], blockRenderPos.x, blockRenderPos.y);
                    }
                }
            });
        
            this.canvasContext.fillStyle = `red`;
            this.canvasContext.beginPath();
            this.canvasContext.moveTo(regionRenderPos.x + (regionWidthHeight*this.renderTileSize), regionRenderPos.y);
            this.canvasContext.lineTo(regionRenderPos.x, regionRenderPos.y + (regionWidthHeight*this.renderTileSize));
            this.canvasContext.stroke();
        });

        this.canvasContext.restore();
    }

    debugRender() {
        this.canvasContext.beginPath();
        this.canvasContext.moveTo(this.canvas.width / 2, 0);
        this.canvasContext.lineTo(this.canvas.width / 2, this.canvas.height);
        this.canvasContext.strokeStyle = 'rgb(0,255,0)'
        this.canvasContext.stroke();

        this.canvasContext.beginPath();
        this.canvasContext.moveTo(0, this.canvas.height / 2);
        this.canvasContext.lineTo(this.canvas.width, this.canvas.height / 2);
        this.canvasContext.strokeStyle = 'rgb(0,0,255)'
        this.canvasContext.stroke();

        this.canvasContext.font = "10px Arial ";
        this.canvasContext.fillStyle = 'rgb(255,0,0)'

        for(let y = -15; y < 15; y += 2) {
            for(let x = -15; x < 15; x += 2) {
               this.canvasContext.fillText(
                    `[${x},${y}]`,
                    (x * this.renderTileSize * 32) + (this.canvas.width/2),
                    (y * this.renderTileSize * 32) + (this.canvas.height/2)
                );
            }
        }
    }
}

window.addEventListener('DOMContentLoaded', () => {
    // @ts-ignore for now
    window.mapHandler = new MapHandler();
})