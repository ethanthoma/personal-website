self.onmessage = (event) => {
    const {
        requestId,
        pixels: pixelBuffer,
        width,
        height,
        seedX,
        seedY,
        tolerance,
        contiguous,
    } = event.data;
    if (!(pixelBuffer instanceof ArrayBuffer)) {
        self.postMessage({
            requestId,
            error: "The selection request did not include a pixel buffer.",
        });
        return;
    }
    if (
        !Number.isInteger(width) ||
        !Number.isInteger(height) ||
        width <= 0 ||
        height <= 0
    ) {
        self.postMessage({
            requestId,
            error: "The selection request had invalid dimensions.",
        });
        return;
    }
    if (pixelBuffer.byteLength !== width * height * 4) {
        self.postMessage({
            requestId,
            error: "The selection buffer did not match the image dimensions.",
        });
        return;
    }
    if (
        !Number.isInteger(seedX) ||
        !Number.isInteger(seedY) ||
        seedX < 0 ||
        seedX >= width ||
        seedY < 0 ||
        seedY >= height
    ) {
        self.postMessage({
            requestId,
            error: "The selection seed was outside the image.",
        });
        return;
    }
    if (!Number.isFinite(tolerance) || tolerance < 0 || tolerance > 255) {
        self.postMessage({
            requestId,
            error: "The selection tolerance was invalid.",
        });
        return;
    }

    const pixels = new Uint8Array(pixelBuffer);
    const mask = new Uint8Array(width * height);
    const seed = (seedY * width + seedX) * 4;
    const targetRed = pixels[seed];
    const targetGreen = pixels[seed + 1];
    const targetBlue = pixels[seed + 2];
    const targetAlpha = pixels[seed + 3];
    const toleranceSquared = tolerance * tolerance * 4;
    const stack = [];
    let count = 0;
    let minX = width;
    let minY = height;
    let maxX = -1;
    let maxY = -1;

    function matches(index) {
        const offset = index * 4;
        const red = pixels[offset] - targetRed;
        const green = pixels[offset + 1] - targetGreen;
        const blue = pixels[offset + 2] - targetBlue;
        const alpha = pixels[offset + 3] - targetAlpha;
        return (
            red * red + green * green + blue * blue + alpha * alpha <=
            toleranceSquared
        );
    }

    function shouldFill(x, y) {
        const index = y * width + x;
        return mask[index] === 0 && matches(index);
    }

    function mark(x, y) {
        mask[y * width + x] = 1;
        count += 1;
        if (x < minX) minX = x;
        if (x > maxX) maxX = x;
        if (y < minY) minY = y;
        if (y > maxY) maxY = y;
    }

    if (contiguous) {
        stack.push(seedX, seedY);
        while (stack.length > 0) {
            const y = stack.pop();
            const x = stack.pop();
            if (y < 0 || y >= height) continue;

            let left = x;
            while (left >= 0 && shouldFill(left, y)) left -= 1;
            left += 1;

            let spanAbove = false;
            let spanBelow = false;
            for (
                let currentX = left;
                currentX < width && shouldFill(currentX, y);
                currentX += 1
            ) {
                mark(currentX, y);

                if (y > 0) {
                    const above = shouldFill(currentX, y - 1);
                    if (above && !spanAbove) {
                        stack.push(currentX, y - 1);
                        spanAbove = true;
                    } else if (!above) {
                        spanAbove = false;
                    }
                }

                if (y + 1 < height) {
                    const below = shouldFill(currentX, y + 1);
                    if (below && !spanBelow) {
                        stack.push(currentX, y + 1);
                        spanBelow = true;
                    } else if (!below) {
                        spanBelow = false;
                    }
                }
            }
        }
    } else {
        for (let y = 0; y < height; y += 1) {
            for (let x = 0; x < width; x += 1) {
                if (shouldFill(x, y)) mark(x, y);
            }
        }
    }

    const bounds =
        count > 0
            ? { minX, minY, maxX, maxY }
            : { minX: 0, minY: 0, maxX: 0, maxY: 0 };
    self.postMessage({ requestId, mask: mask.buffer, bounds, count }, [
        mask.buffer,
    ]);
};
