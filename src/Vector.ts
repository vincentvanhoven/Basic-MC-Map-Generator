export default class Vector {
    public x: number;
    public y: number;

    constructor(x: number, y: number) {
        this.x = x;
        this.y = y;
    }

    subtract(vector: Vector|number) {
        this.x -= typeof vector === 'object' ? vector.x : vector;
        this.y -= typeof vector === 'object' ? vector.y : vector;

        return this;
    }

    add(vector: Vector|number) {
        this.x += typeof vector === 'object' ? vector.x : vector;
        this.y += typeof vector === 'object' ? vector.y : vector;

        return this;
    }

    multiply(vector: Vector|number) {
        this.x *= typeof vector === 'object' ? vector.x : vector;
        this.y *= typeof vector === 'object' ? vector.y : vector;

        return this;
    }
}