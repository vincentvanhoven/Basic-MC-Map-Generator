export default class Matrix {
    public matrix: number[] = [
        // 1 0 0
        // 0 1 0
        // 0 0 1
    ];

    constructor(matrix: number[]) {
        if(matrix.length !== 9) {
            throw Error("Matrixes must be initialized with 9 values.")
        }

        this.matrix = matrix;
    }
}