import './style.css';

import { Region } from "./Region";
import Vector from "./Vector";
import CacheManager from './CacheManager';
import RenderManager from './RenderManager';

class MapHandler {
    private cache: CacheManager = new CacheManager();
    private renderManager: RenderManager = new RenderManager();
    
    private canvas: HTMLCanvasElement;
    private canvasContext: CanvasRenderingContext2D;
    
    private regions: Region[] = [];
    public zoomLevel: number = 1000;
    private isMouseDown: boolean = false;
    private dragOffsetInPx: Vector = new Vector(0, 0);

    constructor() {
        this.canvas = <HTMLCanvasElement>document.getElementById("canvas");
        this.canvas.width = window.innerWidth;
        this.canvas.height = window.innerHeight;

        this.canvasContext = <CanvasRenderingContext2D>this.canvas.getContext("2d");
        this.canvasContext.imageSmoothingEnabled = false;

        this.fetchAllRegions();

        window.setInterval(() => {
            let canvasSize: Vector = new Vector(this.canvas.width, this.canvas.height);
            
            this.renderManager.render(this.canvasContext, canvasSize, this.regions, this.zoomLevel, this.dragOffsetInPx);
        }, 1000/15);
        
        window.addEventListener('resize', this.onResizeEvent.bind(this));
        window.addEventListener('mousedown', this.onMouseDownEvent.bind(this));
        window.addEventListener('mousemove', this.onMouseMoveEvent.bind(this));
        window.addEventListener('mouseup', this.onMouseUpEvent.bind(this));
        document.addEventListener('wheel', this.onWheelEvent.bind(this));
    }

    async fetchAllRegions() {
        await fetch('/regionslist')
            .then((response) => {
                return response.json();
            })
            .then(async (json: {PosX: number, PosZ: number}[]) => {
                this.regions = json.map((region) => {
                    return new Region(region.PosX, region.PosZ)
                });
            });
        
        await fetch('/palette')
            .then((response) => {
                return response.json()
            })
            .then(async (json) => {
                this.renderManager.setPalette(json);
            });

        this.downloadMissingRegionsInViewport();
    }

    onMouseDownEvent(_event: MouseEvent) {
        this.isMouseDown = true;
    }

    onMouseMoveEvent(event: MouseEvent) {
        if(this.isMouseDown) {
            this.dragOffsetInPx.x += Math.round(event.movementX / (this.zoomLevel / 1000));
            this.dragOffsetInPx.y += Math.round(event.movementY / (this.zoomLevel / 1000));
            
            this.downloadMissingRegionsInViewport();
        }
    }

    onMouseUpEvent(_event: MouseEvent) {
        this.isMouseDown = false;
    }

    onWheelEvent(event: WheelEvent) {
        this.zoomLevel -= event.deltaY;

        // Clamp the zoomLevel such that the effective renderTileSize will not have decimals. For example:
        // - Without:
        //   - zoomLevel: 900
        //   - renderTileSize: 16
        //   - effective render tile size: 16 * 0.9 = 14.4
        // - With:
        //   - zoomLevel: 900
        //   - renderTileSize: 16
        //   - clamped zoomLevel: 875
        //   - effective render tile size: 16*0.875 = 14
        this.zoomLevel = (
            Math.round(
                this.renderManager.renderTileSize * (this.zoomLevel / 1000)
            ) * 1000
        ) / 16;

        this.downloadMissingRegionsInViewport();
    }

    private onResizeEvent(_event: Event): void {
        this.canvas.width = window.innerWidth;
        this.canvas.height = window.innerHeight;

        this.downloadMissingRegionsInViewport();
    }

    private downloadMissingRegionsInViewport() {
        let canvasSize: Vector = new Vector(this.canvas.width, this.canvas.height);
        let totalOffset: Vector = this.renderManager.getTotalChunkOffset(this.dragOffsetInPx, canvasSize);

        this.regions.forEach((region: Region) => {
            region.downloadDataIfMissing(canvasSize, this.renderManager.renderTileSize, totalOffset, this.zoomLevel);
        })
    }
}

window.addEventListener('DOMContentLoaded', () => {
    // @ts-ignore for now
    window.mapHandler = new MapHandler();
})
