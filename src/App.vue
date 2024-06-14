<template>
    <img
        v-for="region of regionsToLoad"
        :key="`tile-${region.PosX}-${region.PosZ}`"
        :src="`http://127.0.0.1:8181/region/${region.PosX}/${region.PosZ}/render`"
        class="absolute"
        :style="{
            width: `${tileSize}px`,
            height: `${tileSize}px`,
            position: 'absolute',
            top: `${(region.PosZ * tileSize) + dragOffset.z}px`,
            left: `${(region.PosX * tileSize) + dragOffset.x}px`,
        }"
        @load="onImageLoaded($event, region)"
    >
</template>

<script setup lang="ts">
    import {ref, Ref, ComputedRef, computed, onMounted} from "vue";

    type Region = {
        PosX: number;
        PosZ: number;
    };

    // Data
    const regions: Ref<Region[]> = ref([]);
    const numberOfImagesLoaded: Ref<number> = ref(0);
    const tileSize: Ref<number> = ref(24);
    const dragOffset: Ref<{x: number, z: number}> = ref({x: 0, z: 0});

    // On create
    fetchRegionList();

    // Event listeners
    onMounted(() => {
        dragOffset.value.x = window.innerWidth / 2;
        dragOffset.value.z = window.innerHeight / 2;
    })

    // Computed properties
    const regionsToLoad: ComputedRef<Region[]> = computed(() => {
        return regions.value.slice(0, numberOfImagesLoaded.value + 10);
    });

    // Methods
    async function fetchRegionList() {
        await fetch('/region/list')
            .then((response) => {
                return response.json();
            })
            .then(async (json: {PosX: number, PosZ: number}[]) => {
                regions.value = json.map((region) => {
                    return {
                        PosX: region.PosX,
                        PosZ: region.PosZ,
                    }
                });
            });
    }

    function onImageLoaded(_event: Event, _region: Region): void {
        numberOfImagesLoaded.value++;
    }
</script>

<style lang="scss" scoped>
    
</style>
