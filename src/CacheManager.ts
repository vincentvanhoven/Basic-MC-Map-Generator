export default class CacheManager {
    private cache: {[key:string]: any} = {};

    static instance: CacheManager;

    constructor() {
        if (CacheManager.instance) {
            return CacheManager.instance
        }

        CacheManager.instance = this
    }

    set(key: string, value: any) {
        this.cache[key] = value;
    }

    get(key: string) {
        return this.cache.hasOwnProperty(key) ? this.cache[key] : null;
    }

    remember(key: string, callback: () => any) {
        let value: any = this.get(key);

        if(key === 'chunk-total-offset') {
            // console.log(value);
        }

        if(value === null) {
            value = callback();
            this.set(key, value);
        }

        return value;
    }

    purgeAll() {
        this.cache = {};
    }
}
