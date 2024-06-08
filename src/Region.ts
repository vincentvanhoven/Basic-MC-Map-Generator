import CacheManager from "./CacheManager";
import Vector from "./Vector";

export class Region {
    private cache: CacheManager = new CacheManager();

    public PosX: number;
    public PosZ: number;
    
    private blockStates: number[] = [];
    private isDownloadingData: boolean = false;

    constructor(PosX: number, PosZ: number) {
        this.PosX = PosX;
        this.PosZ = PosZ;
    }

    public render(
        canvasContext: CanvasRenderingContext2D,
        renderTileSize: number,
        canvasSize: Vector,
        zoomLevel: number,
        totalOffset: Vector,
        palette: { [key: number]: string; },
        loadedImages: {[key:string]: HTMLImageElement}
    ) {
        if(!this.blockStates) {
            return;
        }

        if(!this.isInViewport(canvasSize, renderTileSize, totalOffset, zoomLevel)) {
            return;
        }

        let regionRenderSize = this.getRenderSize(renderTileSize);
        let regionRenderPos = this.getRenderPosition(renderTileSize, totalOffset);
        
        this.blockStates.forEach((paletteIndex, blockIndex) => {
            let blockState = palette[paletteIndex];

            if(blockState) {
                if(!loadedImages[blockState]) {
                    loadedImages[blockState] = new Image();
                    loadedImages[blockState].src = `resourcepack/textures/block/${blockState}.png`;
                } else {
                    let blockRenderPos = new Vector(
                        (blockIndex % (regionRenderSize / renderTileSize)) * renderTileSize,
                        Math.floor(blockIndex / (regionRenderSize / renderTileSize)) * renderTileSize,
                    ).add(regionRenderPos);

                    canvasContext.drawImage(loadedImages[blockState], blockRenderPos.x, blockRenderPos.y);
                }
            }
        });
        
        this.debugRender(canvasContext, renderTileSize, totalOffset);
    }

    private debugRender(canvasContext: CanvasRenderingContext2D, renderTileSize: number, totalOffset: Vector) {
        let regionRenderSize = this.getRenderSize(renderTileSize);
        let regionRenderPos = this.getRenderPosition(renderTileSize, totalOffset);

        canvasContext.strokeStyle = `red`;
        canvasContext.lineWidth = 4;
        canvasContext.beginPath();
        
        // Outline
        canvasContext.moveTo(regionRenderPos.x, regionRenderPos.y);
        canvasContext.lineTo(regionRenderPos.x + (regionRenderSize * renderTileSize), regionRenderPos.y);
        canvasContext.lineTo(regionRenderPos.x + (regionRenderSize * renderTileSize), regionRenderPos.y + (regionRenderSize * renderTileSize));
        canvasContext.lineTo(regionRenderPos.x, regionRenderPos.y + (regionRenderSize * renderTileSize));
        
        // Diagonals
        canvasContext.lineTo(regionRenderPos.x + (regionRenderSize * renderTileSize), regionRenderPos.y);
        canvasContext.moveTo(regionRenderPos.x, regionRenderPos.y);
        canvasContext.lineTo(regionRenderPos.x + (regionRenderSize * renderTileSize), regionRenderPos.y + (regionRenderSize * renderTileSize));

        canvasContext.stroke();
    }

    public async downloadDataIfMissing(canvasSize: Vector, renderTileSize: number, totalOffset: Vector, zoomLevel: number): Promise<boolean> {
        if(
            !this.isDownloadingData
            && this.blockStates.length == 0
            && this.isInViewport(canvasSize, renderTileSize, totalOffset, zoomLevel)
        ) {
            this.isDownloadingData = true;

            return fetch(`/blockdata?region_x=${this.PosX}&region_z=${this.PosZ}`)
                .then((response) => {
                    return response.json();
                })
                .then((json) => {
                    this.isDownloadingData = false;
                    this.blockStates = json;
                    return true;
                })
                .catch((err) => {
                    this.isDownloadingData = false;
                    throw err;
                });
        }

        return false;
    }

    private getRenderSize(renderTileSize: number) {
        return this.cache.remember(`region-render-size-${this.PosX}-${this.PosZ}`, () => {
            let regionWidthHeight = 32*16;
            return regionWidthHeight * renderTileSize;
        })
    }

    private getRenderPosition(renderTileSize: number, totalOffset: Vector) {
        return this.cache.remember(`region-render-position-${this.PosX}-${this.PosZ}`, () => {
            let regionRenderSize = this.getRenderSize(renderTileSize);

            let regionRenderPos: Vector = new Vector(this.PosX, this.PosZ);

            return regionRenderPos
                .multiply(regionRenderSize)
                .add(totalOffset);
        });
    }

    private isInViewport(canvasSize: Vector, renderTileSize: number, totalOffset: Vector, zoomLevel: number) {
        return this.cache.remember(`region-isInViewport-${this.PosX}-${this.PosZ}`, () => {
            let regionRenderSize = this.getRenderSize(renderTileSize);
            let regionRenderPos = this.getRenderPosition(renderTileSize, totalOffset);

            let scaledCanvasSize: Vector = (new Vector(canvasSize.x, canvasSize.y)).divide(zoomLevel / 1000);

            return regionRenderPos.x <= scaledCanvasSize.x
                && (regionRenderPos.x + regionRenderSize) >= 0
                && regionRenderPos.y <= scaledCanvasSize.y
                && (regionRenderPos.y + regionRenderSize) >= 0;
        });
    }
}