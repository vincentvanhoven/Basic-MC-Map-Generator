import CacheManager from "./CacheManager";
import { Region } from "./Region";
import Vector from "./Vector";


export default class RenderManager {
    private cache: CacheManager = new CacheManager();

    private loadedImages: {[key:string]: HTMLImageElement} = {};
    private palette: {[key:number]: string} = [];

    public renderTileSize: number = 16;
    
    render(canvasContext: CanvasRenderingContext2D, canvasSize: Vector, regions: Region[], zoomLevel: number, dragOffsetInPx: Vector): void {
        console.time("render time");

        canvasContext.save();

        canvasContext.fillStyle = `black`;
        canvasContext.fillRect(0, 0, canvasSize.x, canvasSize.y);

        // Scale the entire canvas by the current zoom level
        canvasContext.scale(zoomLevel / 1000, zoomLevel / 1000)

        regions.forEach((region: Region) => {
            region.render(
                canvasContext,
                this.renderTileSize,
                canvasSize,
                zoomLevel,
                // Determine compensation for keeping negative coords inside viewport.
                this.getTotalChunkOffset(dragOffsetInPx, canvasSize),
                this.palette,
                this.loadedImages,
            );
        });

        canvasContext.restore();

        console.timeEnd("render time");
    }

    public setPalette(palette: { [key: number]: string; }): void {
        this.palette = palette;
    }    

    public getCanvasCenterOffset(canvasSize: Vector): Vector {
        return new Vector(
            Math.round(canvasSize.x / 2),
            Math.round(canvasSize.y / 2),
        );
    }

    public getTotalChunkOffset(dragOffsetInPx: Vector, canvasSize: Vector): Vector {
        return new Vector(
            this.getCanvasCenterOffset(canvasSize).x + dragOffsetInPx.x,
            this.getCanvasCenterOffset(canvasSize).y + dragOffsetInPx.y,
        );
    }
}