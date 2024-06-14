<template>
    <div
        class="absolute"
        :style="{
            top: `${dragOffset.z}px`,
            left: `${dragOffset.x}px`,
        }"
    >
        <div
            v-for="(region, index) of regions"
            :key="`tile-${region.PosX}-${region.PosZ}`"
        >
            <template v-if="index <= numberOfImagesLoaded + 10">
                <img
                    v-show="region.isLoaded"
                    :src="`http://127.0.0.1:8181/region/${region.PosX}/${region.PosZ}/render`"
                    class="absolute max-w-none"
                    :style="{
                        width: `${tileSize}px`,
                        height: `${tileSize}px`,
                        top: `${(region.PosZ * tileSize)}px`,
                        left: `${(region.PosX * tileSize)}px`,
                    }"
                    @load="onImageLoaded($event, region)"
                >
            </template>

            <span
                v-if="index > numberOfImagesLoaded + 10 || !region.isLoaded"
                class="absolute bg-gray-300 animate-pulse"
                :style="{
                    width: `${tileSize}px`,
                    height: `${tileSize}px`,
                    top: `${(region.PosZ * tileSize)}px`,
                    left: `${(region.PosX * tileSize)}px`,
                }"
            ></span>
        </div>
    </div>
</template>

<script setup lang="ts">
    import {ref, Ref, ComputedRef, computed, onMounted, nextTick, onUnmounted} from "vue";

    type Region = {
        PosX: number;
        PosZ: number;
        isLoaded: boolean;
    };

    // Data
    const regions: Ref<Region[]> = ref([]);
    const numberOfImagesLoaded: Ref<number> = ref(0);
    const tileSize: Ref<number> = ref(24);
    const dragOffset: Ref<{x: number, z: number}> = ref({x: 0, z: 0});
    const mouseIsDown: Ref<boolean> = ref(false);

    // Event listeners
    onMounted(() => {
        dragOffset.value.x = window.innerWidth / 2;
        dragOffset.value.z = window.innerHeight / 2;

        if(document.readyState === "complete") {
            fetchRegionList();
        } else {
            document.addEventListener("DOMContentLoaded", () => {
                fetchRegionList();
            })
        }

        window.addEventListener('mousedown', onMouseDown);
        window.addEventListener('mouseup', onMouseUp);
        window.addEventListener('mouseleave', onMouseUp);
        window.addEventListener('mousemove', onMouseMove);
    });

    // Methods
    async function fetchRegionList() {
        // Sleep 200ms to prevent the browser from staying in DOM 'loading' state while the images are being fetched
        await new Promise(resolve => setTimeout(resolve, 200));

        await fetch('/region/list')
            .then((response) => {
                return response.json();
            })
            .then(async (json: {PosX: number, PosZ: number}[]) => {
                regions.value = json.map((region) => {
                    return {
                        PosX: region.PosX,
                        PosZ: region.PosZ,
                        isLoaded: false,
                    }
                });
            });
    }

    function onImageLoaded(_event: Event, region: Region): void {
        numberOfImagesLoaded.value++;
        region.isLoaded = true;
    }

    function onMouseDown(_event: MouseEvent) {
        mouseIsDown.value = true;
    }

    function onMouseUp(_event: MouseEvent) {
        mouseIsDown.value = false;
    }
    
    function onMouseMove(event: MouseEvent) {
        if(mouseIsDown.value) {
            dragOffset.value.x += event.movementX;
            dragOffset.value.z += event.movementY;
        }
    }
</script>

<style lang="scss" scoped>
    
</style>
