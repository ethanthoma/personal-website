(() => {
    const root = document.getElementById("image-editor-root");
    if (!root) return;

    const canvas = document.getElementById("editor-canvas");
    const dropZone = document.getElementById("editor-drop-zone");
    const status = document.getElementById("editor-status");
    const pasteButton = document.getElementById("editor-paste");
    const fileInput = document.getElementById("editor-file");
    const undoButton = document.getElementById("editor-undo");
    const clearButton = document.getElementById("editor-clear");
    const copyButton = document.getElementById("editor-copy");
    const copyLabel = document.getElementById("editor-copy-label");
    const layerList = document.getElementById("editor-layer-list");
    const layerCount = document.getElementById("editor-layer-count");
    const layerPanel = document.getElementById("editor-layers");
    const layerBackdrop = document.getElementById("editor-layer-backdrop");
    const layerCloseButton = document.getElementById("editor-layer-close");
    const layerRowTemplate = document.getElementById(
        "editor-layer-row-template",
    );
    const layerEmptyTemplate = document.getElementById(
        "editor-layer-empty-template",
    );
    const layerBackwardButton = document.getElementById(
        "editor-layer-backward",
    );
    const layerForwardButton = document.getElementById("editor-layer-forward");
    const layerNewButton = document.getElementById("editor-layer-new");
    const layerDuplicateButton = document.getElementById(
        "editor-layer-duplicate",
    );
    const layerDeleteButton = document.getElementById("editor-layer-delete");
    const fillButton = document.getElementById("editor-fill");
    const clearSelectionButton = document.getElementById(
        "editor-clear-selection",
    );
    const colorInput = document.getElementById("editor-color");
    const colorControl = document.getElementById("editor-color-control");
    const colorToggle = document.getElementById("editor-color-toggle");
    const colorPopover = document.getElementById("editor-color-popover");
    const colorDone = document.getElementById("editor-color-done");
    const colorSwatch = document.getElementById("editor-color-swatch");
    const colorSwatchContext = colorSwatch.getContext("2d");
    const colorPlane = document.getElementById("editor-color-plane");
    const colorPlaneContext = colorPlane.getContext("2d");
    const colorHue = document.getElementById("editor-color-hue");
    const colorHueContext = colorHue.getContext("2d");
    const colorHex = document.getElementById("editor-color-hex");
    const colorPresets = Array.from(
        root.querySelectorAll("[data-editor-color]"),
    );
    const widthInput = document.getElementById("editor-width");
    const widthValue = document.getElementById("editor-width-value");
    const toleranceInput = document.getElementById("editor-tolerance");
    const toleranceValue = document.getElementById("editor-tolerance-value");
    const contiguousInput = document.getElementById("editor-contiguous");
    const modeReadout = document.getElementById("editor-mode-readout");
    const frameReadout = document.getElementById("editor-frame-readout");
    const layerReadout = document.getElementById("editor-layer-readout");
    const cursorReadout = document.getElementById("editor-cursor-readout");
    const historyReadout = document.getElementById("editor-history-readout");
    const mobileAddButton = document.getElementById("editor-mobile-add");
    const emptyState = document.getElementById("editor-empty-state");
    const emptyAddButton = document.getElementById("editor-empty-add");
    const mobileLayersButton = document.getElementById("editor-mobile-layers");
    const mobileCopyButton = document.getElementById("editor-mobile-copy");
    const mobileLayerCount = document.getElementById(
        "editor-mobile-layer-count",
    );
    const mobileMode = document.getElementById("editor-mobile-mode");
    const mobileLayer = document.getElementById("editor-mobile-layer");
    const selectGroup = document.getElementById("editor-select-group");
    const selectSummary = document.getElementById("editor-select-summary");
    const selectXInput = document.getElementById("editor-select-x");
    const selectYInput = document.getElementById("editor-select-y");
    const selectWidthInput = document.getElementById("editor-select-width");
    const selectHeightInput = document.getElementById("editor-select-height");
    const selectAspectButton = document.getElementById("editor-select-aspect");
    const selectCenterButton = document.getElementById("editor-select-center");
    const selectFitButton = document.getElementById("editor-select-fit");
    const selectFlipHorizontalButton = document.getElementById(
        "editor-select-flip-horizontal",
    );
    const selectFlipVerticalButton = document.getElementById(
        "editor-select-flip-vertical",
    );
    const selectResetAspectButton = document.getElementById(
        "editor-select-reset-aspect",
    );
    const inkGroup = document.getElementById("editor-ink-group");
    const scanGroup = document.getElementById("editor-scan-group");
    const toolButtons = Array.from(root.querySelectorAll("[data-editor-tool]"));

    const state = {
        board: null,
        draftStroke: null,
        imageInteraction: null,
        loadId: 0,
        pastePoint: null,
        selectionRequestId: 0,
        selectionWorker: null,
        selectAspectLocked: false,
        tool: "select",
        ready: false,
    };

    const DEFAULT_CANVAS_WIDTH = 1280;
    const DEFAULT_CANVAS_HEIGHT = 720;
    const PORTRAIT_CANVAS_HEIGHT = 1440;
    const PORTRAIT_ASPECT_MIN = 0.62;
    const PORTRAIT_ASPECT_MAX = 0.9;
    const PASTE_OFFSET_PIXELS = 16;
    const TILE_SIZE = 256;
    const HISTORY_BYTES_MAX = 64 * 1024 * 1024;
    const HISTORY_ENTRY_MAX = 256;
    const PENCIL_POINTS_MAX = 8192;
    const IMAGE_MIN_SIZE = 24;
    const IMAGE_MIN_VISIBLE = 24;
    const IMAGE_HANDLE_SIZE_CSS = 10;
    const IMAGE_HANDLE_HIT_CSS = 20;
    const IMAGE_HANDLE_CURSORS = {
        nw: "nwse-resize",
        n: "ns-resize",
        ne: "nesw-resize",
        e: "ew-resize",
        se: "nwse-resize",
        s: "ns-resize",
        sw: "nesw-resize",
        w: "ew-resize",
    };
    let initialization = null;
    let pickerHsv = null;
    let colorPlanePointerId = null;
    let colorHuePointerId = null;

    function setStatus(message) {
        status.textContent = message;
    }

    function setLayerSheetOpen(open) {
        layerPanel.dataset.sheetOpen = String(open);
        layerBackdrop.hidden = !open;
        mobileLayersButton.setAttribute("aria-expanded", String(open));
        if (open) setColorPickerOpen(false);
    }

    function toolLabel(tool) {
        if (tool === "wand") return "Magic wand";
        return tool.charAt(0).toUpperCase() + tool.slice(1);
    }

    function activeBoard() {
        return state.board;
    }

    function layerById(board, layerId) {
        if (!board || layerId === null) return null;
        return board.layers.find((layer) => layer.id === layerId) || null;
    }

    function activeLayer() {
        const board = activeBoard();
        if (!board) return null;
        return layerById(board, board.activeLayerId);
    }

    function targetImage() {
        const layer = activeLayer();
        return layer?.kind === "image" ? layer : null;
    }

    function createCanvas(width, height) {
        if (typeof OffscreenCanvas === "function")
            return new OffscreenCanvas(width, height);

        const surface = document.createElement("canvas");
        surface.width = width;
        surface.height = height;
        return surface;
    }

    function createRaster(width, height, fillWhite) {
        assertRasterSize(width, height);
        const surface = createCanvas(width, height);
        const context = surface.getContext("2d", { willReadFrequently: true });
        if (!context) throw new Error("2D canvas unavailable");

        if (fillWhite) {
            context.fillStyle = "#ffffff";
            context.fillRect(0, 0, width, height);
        }
        return { surface, context };
    }

    function assertRasterSize(width, height) {
        if (!Number.isInteger(width))
            throw new Error("Raster width must be an integer");
        if (!Number.isInteger(height))
            throw new Error("Raster height must be an integer");
        if (width <= 0) throw new Error("Raster width must be positive");
        if (height <= 0) throw new Error("Raster height must be positive");
    }

    function createBoard(width, height) {
        const background = createRaster(width, height, true);
        return {
            width,
            height,
            background: background.surface,
            backgroundContext: background.context,
            layers: [],
            selection: null,
            activeLayerId: null,
            history: [],
            historyBytes: 0,
            nextLayerId: 1,
            nextImageNumber: 1,
            nextInkNumber: 1,
        };
    }

    function createInkSurface(width, height) {
        return createRaster(width, height, false);
    }

    function createImageLayer(board, bitmap, placement) {
        const ink = createInkSurface(bitmap.width, bitmap.height);
        return {
            id: board.nextLayerId++,
            kind: "image",
            name: `Image ${board.nextImageNumber++}`,
            bitmap,
            x: placement.x,
            y: placement.y,
            width: placement.width,
            height: placement.height,
            flipX: false,
            flipY: false,
            contentWidth: bitmap.width,
            contentHeight: bitmap.height,
            ink: ink.surface,
            inkContext: ink.context,
            inkOps: 0,
            visible: true,
        };
    }

    function createInkLayer(board) {
        const ink = createInkSurface(board.width, board.height);
        return {
            id: board.nextLayerId++,
            kind: "ink",
            name: `Ink ${board.nextInkNumber++}`,
            bitmap: null,
            x: 0,
            y: 0,
            width: board.width,
            height: board.height,
            flipX: false,
            flipY: false,
            contentWidth: board.width,
            contentHeight: board.height,
            ink: ink.surface,
            inkContext: ink.context,
            inkOps: 0,
            visible: true,
        };
    }

    function blankBoard() {
        const dimensions = initialBoardDimensions();
        return createBoard(dimensions.width, dimensions.height);
    }

    function initialBoardDimensions() {
        const clientAspect = window.innerWidth / window.innerHeight;
        if (clientAspect >= 1) {
            return {
                width: DEFAULT_CANVAS_WIDTH,
                height: DEFAULT_CANVAS_HEIGHT,
            };
        }

        const aspect = Math.min(
            PORTRAIT_ASPECT_MAX,
            Math.max(PORTRAIT_ASPECT_MIN, clientAspect),
        );
        return {
            width: Math.round((PORTRAIT_CANVAS_HEIGHT * aspect) / 2) * 2,
            height: PORTRAIT_CANVAS_HEIGHT,
        };
    }

    function syncToolContext() {
        const usesSelect = state.tool === "select";
        const usesInk = state.tool === "pencil" || state.tool === "line";
        const usesScan = state.tool === "wand";
        selectGroup.dataset.active = String(usesSelect);
        inkGroup.dataset.active = String(usesInk);
        scanGroup.dataset.active = String(usesScan);
    }

    function updateSelectGeometry(image) {
        const fields = [
            selectXInput,
            selectYInput,
            selectWidthInput,
            selectHeightInput,
        ];
        for (const field of fields) field.disabled = !image;
        selectAspectButton.disabled = !image;
        selectAspectButton.setAttribute(
            "aria-pressed",
            String(state.selectAspectLocked),
        );
        selectAspectButton.title = state.selectAspectLocked
            ? "Unlock aspect ratio"
            : "Lock aspect ratio";
        selectAspectButton.setAttribute("aria-label", selectAspectButton.title);
        selectCenterButton.disabled = !image;
        selectFitButton.disabled = !image;
        selectFlipHorizontalButton.disabled = !image;
        selectFlipVerticalButton.disabled = !image;
        selectResetAspectButton.disabled = !image;
        if (!image) {
            for (const field of fields) field.value = "";
            return;
        }
        selectXInput.value = String(Math.round(image.x));
        selectYInput.value = String(Math.round(image.y));
        selectWidthInput.value = String(Math.round(image.width));
        selectHeightInput.value = String(Math.round(image.height));
    }

    function updateControls() {
        const board = activeBoard();
        if (!board) return;

        const layer = activeLayer();
        const selectedImage = targetImage();
        const hasInk = Boolean(layer && layer.inkOps > 0);
        const hasSelection = Boolean(
            board.selection && board.selection.count > 0,
        );
        undoButton.disabled = board.history.length === 0;
        clearButton.disabled = !hasInk;
        copyButton.disabled = !state.ready;
        mobileCopyButton.disabled = !state.ready;
        selectSummary.textContent = selectedImage
            ? `${selectedImage.name} · ${Math.round(selectedImage.width)} × ${Math.round(selectedImage.height)} · drag to move`
            : layer
              ? `${layer.name} · full canvas · draw directly onto this layer`
              : "Paste an image or draw to create a layer.";
        updateSelectGeometry(selectedImage);
        const copyTarget = selectedImage?.visible ? selectedImage : null;
        const copyLabelText = copyTarget ? "Copy image" : "Copy canvas";
        copyLabel.textContent = copyLabelText;
        copyButton.setAttribute("aria-label", copyLabelText);
        copyButton.title = copyLabelText;
        mobileCopyButton.setAttribute("aria-label", copyLabelText);
        mobileCopyButton.title = copyLabelText;
        fillButton.disabled = !hasSelection || !layer?.visible;
        clearSelectionButton.disabled = !hasSelection;
        const selectedIndex = layer ? board.layers.indexOf(layer) : -1;
        layerBackwardButton.disabled = selectedIndex <= 0;
        layerForwardButton.disabled =
            selectedIndex < 0 || selectedIndex === board.layers.length - 1;
        layerDuplicateButton.disabled = !layer;
        layerDeleteButton.disabled = !layer;
        emptyState.hidden = board.layers.length > 0 || state.tool !== "select";
        mobileLayerCount.hidden = board.layers.length === 0;
        mobileLayersButton.setAttribute(
            "aria-label",
            board.layers.length === 0
                ? "Open layers"
                : `Open layers, ${board.layers.length} total`,
        );
        modeReadout.textContent = state.tool.toUpperCase();
        mobileMode.textContent = toolLabel(state.tool);
        frameReadout.textContent = `${board.width} × ${board.height}`;
        layerReadout.textContent = layer ? layer.name : "—";
        layerReadout.title = layer ? layer.name : "No active layer";
        mobileLayer.textContent = layer ? layer.name : "No layer";
        mobileLayer.title = layer ? layer.name : "No active layer";
        historyReadout.textContent = String(board.history.length).padStart(
            2,
            "0",
        );
        syncToolContext();
        renderLayerPanel(board);

        if (state.tool === "wand") canvas.style.cursor = "cell";
        else if (state.tool === "select") canvas.style.cursor = "default";
        else canvas.style.cursor = "crosshair";

        for (const button of toolButtons) {
            button.setAttribute(
                "aria-pressed",
                String(button.dataset.editorTool === state.tool),
            );
        }
    }

    function updateCursorReadout(event) {
        const point = canvasPoint(event);
        const text = `${String(point.x).padStart(4, "0")} / ${String(point.y).padStart(4, "0")}`;
        cursorReadout.textContent = text;
    }

    function blankStatus(board) {
        return `Blank ${board.width} × ${board.height}`;
    }

    function updateWidthLabel() {
        widthValue.value = `${widthInput.value} px`;
        widthValue.textContent = `${widthInput.value} px`;
    }

    function clampUnit(value) {
        return Math.min(1, Math.max(0, value));
    }

    function normalizeHex(value) {
        const match = /^#?([0-9a-f]{6})$/i.exec(value.trim());
        return match ? `#${match[1].toLowerCase()}` : null;
    }

    function hsvFromHex(value) {
        const normalized = normalizeHex(value);
        if (!normalized) return null;

        const packed = Number.parseInt(normalized.slice(1), 16);
        const red = ((packed >> 16) & 255) / 255;
        const green = ((packed >> 8) & 255) / 255;
        const blue = (packed & 255) / 255;
        const maximum = Math.max(red, green, blue);
        const minimum = Math.min(red, green, blue);
        const delta = maximum - minimum;
        let hue = 0;

        if (delta > 0) {
            if (maximum === red) hue = ((green - blue) / delta) % 6;
            else if (maximum === green) hue = (blue - red) / delta + 2;
            else hue = (red - green) / delta + 4;
            hue *= 60;
            if (hue < 0) hue += 360;
        }

        return {
            h: hue,
            s: maximum === 0 ? 0 : delta / maximum,
            v: maximum,
        };
    }

    function hsvToHex(hsv) {
        const hue = ((hsv.h % 360) + 360) % 360;
        const saturation = clampUnit(hsv.s);
        const value = clampUnit(hsv.v);
        const chroma = value * saturation;
        const section = hue / 60;
        const intermediate = chroma * (1 - Math.abs((section % 2) - 1));
        let red = 0;
        let green = 0;
        let blue = 0;

        if (section < 1) {
            red = chroma;
            green = intermediate;
        } else if (section < 2) {
            red = intermediate;
            green = chroma;
        } else if (section < 3) {
            green = chroma;
            blue = intermediate;
        } else if (section < 4) {
            green = intermediate;
            blue = chroma;
        } else if (section < 5) {
            red = intermediate;
            blue = chroma;
        } else {
            red = chroma;
            blue = intermediate;
        }

        const offset = value - chroma;
        const channelHex = (channel) =>
            Math.round((channel + offset) * 255)
                .toString(16)
                .padStart(2, "0");
        return `#${channelHex(red)}${channelHex(green)}${channelHex(blue)}`;
    }

    function updateColorSwatch() {
        if (!colorSwatchContext) return;
        colorSwatchContext.fillStyle = colorInput.value;
        colorSwatchContext.fillRect(
            0,
            0,
            colorSwatch.width,
            colorSwatch.height,
        );
    }

    function updateColorPresets() {
        for (const preset of colorPresets) {
            preset.setAttribute(
                "aria-pressed",
                String(preset.dataset.editorColor === colorInput.value),
            );
        }
    }

    function drawColorPlane() {
        if (!colorPlaneContext || !pickerHsv) return;
        const width = colorPlane.width;
        const height = colorPlane.height;
        colorPlaneContext.fillStyle = `hsl(${pickerHsv.h}, 100%, 50%)`;
        colorPlaneContext.fillRect(0, 0, width, height);

        const white = colorPlaneContext.createLinearGradient(0, 0, width, 0);
        white.addColorStop(0, "rgba(255, 255, 255, 1)");
        white.addColorStop(1, "rgba(255, 255, 255, 0)");
        colorPlaneContext.fillStyle = white;
        colorPlaneContext.fillRect(0, 0, width, height);

        const black = colorPlaneContext.createLinearGradient(0, 0, 0, height);
        black.addColorStop(0, "rgba(0, 0, 0, 0)");
        black.addColorStop(1, "rgba(0, 0, 0, 1)");
        colorPlaneContext.fillStyle = black;
        colorPlaneContext.fillRect(0, 0, width, height);

        const x = clampUnit(pickerHsv.s) * width;
        const y = (1 - clampUnit(pickerHsv.v)) * height;
        colorPlaneContext.beginPath();
        colorPlaneContext.arc(x, y, 5, 0, Math.PI * 2);
        colorPlaneContext.lineWidth = 3;
        colorPlaneContext.strokeStyle = "#141413";
        colorPlaneContext.stroke();
        colorPlaneContext.beginPath();
        colorPlaneContext.arc(x, y, 3, 0, Math.PI * 2);
        colorPlaneContext.lineWidth = 1;
        colorPlaneContext.strokeStyle = "#fbf8ef";
        colorPlaneContext.stroke();
        colorPlane.setAttribute(
            "aria-label",
            `Saturation ${Math.round(pickerHsv.s * 100)}%, brightness ${Math.round(pickerHsv.v * 100)}%. Use the arrow keys to adjust.`,
        );
    }

    function drawColorHue() {
        if (!colorHueContext || !pickerHsv) return;
        const width = colorHue.width;
        const height = colorHue.height;
        const gradient = colorHueContext.createLinearGradient(0, 0, width, 0);
        for (const [offset, colour] of [
            [0, "#ff0000"],
            [1 / 6, "#ffff00"],
            [2 / 6, "#00ff00"],
            [3 / 6, "#00ffff"],
            [4 / 6, "#0000ff"],
            [5 / 6, "#ff00ff"],
            [1, "#ff0000"],
        ]) {
            gradient.addColorStop(offset, colour);
        }
        colorHueContext.fillStyle = gradient;
        colorHueContext.fillRect(0, 0, width, height);

        const x = (pickerHsv.h / 360) * width;
        for (const [strokeStyle, lineWidth] of [
            ["#141413", 4],
            ["#fbf8ef", 2],
        ]) {
            colorHueContext.beginPath();
            colorHueContext.moveTo(x, 0);
            colorHueContext.lineTo(x, height);
            colorHueContext.lineWidth = lineWidth;
            colorHueContext.strokeStyle = strokeStyle;
            colorHueContext.stroke();
        }
        colorHue.setAttribute("aria-valuenow", String(Math.round(pickerHsv.h)));
    }

    function drawColorPicker() {
        drawColorPlane();
        drawColorHue();
    }

    function setInkColor(value, hsv = hsvFromHex(value)) {
        const normalized = normalizeHex(value);
        if (!normalized || !hsv) return;
        colorInput.value = normalized;
        pickerHsv = hsv;
        colorHex.value = normalized.toUpperCase();
        colorHex.removeAttribute("aria-invalid");
        updateColorSwatch();
        updateColorPresets();
        drawColorPicker();
    }

    function setColorPickerOpen(open) {
        if (open) setLayerSheetOpen(false);
        colorPopover.hidden = !open;
        colorToggle.setAttribute("aria-expanded", String(open));
        if (open) drawColorPicker();
    }

    function pointerFraction(event, element) {
        const rect = element.getBoundingClientRect();
        if (rect.width === 0 || rect.height === 0) return null;
        return {
            x: clampUnit((event.clientX - rect.left) / rect.width),
            y: clampUnit((event.clientY - rect.top) / rect.height),
        };
    }

    function updatePlaneFromPointer(event) {
        if (!pickerHsv) return;
        const point = pointerFraction(event, colorPlane);
        if (!point) return;
        const hsv = { h: pickerHsv.h, s: point.x, v: 1 - point.y };
        setInkColor(hsvToHex(hsv), hsv);
    }

    function updateHueFromPointer(event) {
        if (!pickerHsv) return;
        const point = pointerFraction(event, colorHue);
        if (!point) return;
        const hsv = {
            ...pickerHsv,
            h: Math.min(359.999, point.x * 360),
        };
        setInkColor(hsvToHex(hsv), hsv);
    }

    function applyHexInput() {
        const value = normalizeHex(colorHex.value);
        if (!value) {
            colorHex.setAttribute("aria-invalid", "true");
            return;
        }
        setInkColor(value);
    }

    function restoreHexInput() {
        if (normalizeHex(colorHex.value)) {
            applyHexInput();
            return;
        }
        colorHex.value = colorInput.value.toUpperCase();
        colorHex.removeAttribute("aria-invalid");
    }

    function updateToleranceLabel() {
        toleranceValue.value = `${toleranceInput.value}%`;
        toleranceValue.textContent = `${toleranceInput.value}%`;
    }

    async function decodeImage(blob, maxWidth, maxHeight) {
        const bitmap = await createImageBitmap(blob, {
            imageOrientation: "from-image",
        });
        const scale = Math.min(
            1,
            maxWidth / bitmap.width,
            maxHeight / bitmap.height,
        );
        if (scale >= 1) return bitmap;

        try {
            return await createImageBitmap(bitmap, {
                resizeWidth: Math.max(1, Math.round(bitmap.width * scale)),
                resizeHeight: Math.max(1, Math.round(bitmap.height * scale)),
                resizeQuality: "high",
            });
        } finally {
            bitmap.close();
        }
    }

    function closeBitmaps(bitmaps) {
        for (const bitmap of bitmaps) bitmap.close();
    }

    function pasteOrigin(bitmap, index) {
        const board = activeBoard();
        const width = board.width;
        const height = board.height;
        const anchor = state.pastePoint || {
            x: width / 2,
            y: height / 2,
        };
        const offset = index * PASTE_OFFSET_PIXELS;
        return {
            x: Math.max(
                0,
                Math.min(
                    width - bitmap.width,
                    Math.round(anchor.x - bitmap.width / 2 + offset),
                ),
            ),
            y: Math.max(
                0,
                Math.min(
                    height - bitmap.height,
                    Math.round(anchor.y - bitmap.height / 2 + offset),
                ),
            ),
            width: bitmap.width,
            height: bitmap.height,
        };
    }

    async function loadImages(blobs) {
        const images = Array.from(blobs || []).filter(
            (blob) => blob instanceof Blob && blob.type.startsWith("image/"),
        );
        if (images.length === 0) {
            setStatus("No images found. Use the Paste or Add images buttons.");
            return;
        }

        const loadId = ++state.loadId;
        setStatus(
            images.length === 1
                ? "Decoding image…"
                : `Decoding ${images.length} images…`,
        );
        await initialization;
        if (!state.ready) return;
        if (loadId !== state.loadId) return;

        const board = activeBoard();
        const width = board.width;
        const height = board.height;
        const loaded = [];
        for (const blob of images) {
            let bitmap;
            try {
                bitmap = await decodeImage(blob, width, height);
            } catch {
                continue;
            }
            if (loadId !== state.loadId) {
                bitmap.close();
                closeBitmaps(loaded);
                return;
            }
            loaded.push(bitmap);
        }

        if (loaded.length === 0) {
            setStatus("Could not paste those images. They may be unsupported.");
            return;
        }

        const placements = loaded.map((bitmap, index) => {
            return { bitmap, ...pasteOrigin(bitmap, index) };
        });
        const pastedLayers = placements.map((placement) =>
            createImageLayer(board, placement.bitmap, placement),
        );
        board.layers.push(...pastedLayers);
        state.selectionRequestId += 1;
        board.selection = null;
        board.activeLayerId = pastedLayers.at(-1).id;
        state.tool = "select";
        const bytes = pastedLayers.reduce((total, image) => {
            return total + image.width * image.height * 4;
        }, 0);
        addHistory(board, {
            type: "add-images",
            layers: pastedLayers,
            bytes,
        });
        updateControls();
        render();
        if (placements.length === 1) {
            setStatus(
                "Pasted the image. Drag it to move, or drag a handle to resize it.",
            );
        } else {
            setStatus(
                `Pasted ${placements.length} images onto the canvas. Drag the selected one to move it.`,
            );
        }
    }

    function pointerPosition(event) {
        const rect = canvas.getBoundingClientRect();
        return {
            x: ((event.clientX - rect.left) * canvas.width) / rect.width,
            y: ((event.clientY - rect.top) * canvas.height) / rect.height,
        };
    }

    function canvasPoint(event) {
        const point = pointerPosition(event);
        return {
            x: Math.max(0, Math.min(canvas.width - 1, Math.floor(point.x))),
            y: Math.max(0, Math.min(canvas.height - 1, Math.floor(point.y))),
        };
    }

    function imageLineWidth() {
        const rect = canvas.getBoundingClientRect();
        const scale = rect.width > 0 ? canvas.width / rect.width : 1;
        return Math.max(1, Number(widthInput.value) * scale);
    }

    function pointInsideLayer(layer, point) {
        return (
            point.x >= layer.x &&
            point.x <= layer.x + layer.width &&
            point.y >= layer.y &&
            point.y <= layer.y + layer.height
        );
    }

    function layerPoint(layer, point) {
        const localX = ((point.x - layer.x) * layer.contentWidth) / layer.width;
        const localY =
            ((point.y - layer.y) * layer.contentHeight) / layer.height;
        return {
            x: layer.flipX ? layer.contentWidth - localX : localX,
            y: layer.flipY ? layer.contentHeight - localY : localY,
        };
    }

    function layerPointOnCanvas(layer, point) {
        const localX = layer.flipX ? layer.contentWidth - point.x : point.x;
        const localY = layer.flipY ? layer.contentHeight - point.y : point.y;
        return {
            x: layer.x + (localX * layer.width) / layer.contentWidth,
            y: layer.y + (localY * layer.height) / layer.contentHeight,
        };
    }

    function layerLineWidth(layer) {
        return imageLineWidth() * (layer.contentWidth / layer.width);
    }

    function ensureStrokeLayer() {
        const current = activeLayer();
        if (current) return { layer: current, createdLayerId: null };
        const board = activeBoard();
        const layer = createInkLayer(board);
        board.layers.push(layer);
        board.activeLayerId = layer.id;
        return { layer, createdLayerId: layer.id };
    }

    function beginStroke(event) {
        if (!state.ready || event.button !== 0) return;
        if (state.tool !== "pencil" && state.tool !== "line") return;

        event.preventDefault();
        const board = activeBoard();
        const point = pointerPosition(event);
        const current = activeLayer();
        if (current && !current.visible) {
            setStatus("Show the active layer before drawing on it.");
            return;
        }
        if (current && !pointInsideLayer(current, point)) {
            setStatus("Start the stroke inside the active image layer.");
            return;
        }

        canvas.setPointerCapture(event.pointerId);
        const { layer, createdLayerId } = ensureStrokeLayer();
        board.selection = null;
        state.selectionRequestId += 1;
        state.imageInteraction = null;
        state.draftStroke = {
            layerId: layer.id,
            createdLayerId,
            tool: state.tool,
            points: [layerPoint(layer, point)],
            color: colorInput.value,
            width: layerLineWidth(layer),
        };
        updateControls();
        render();
    }

    function appendStrokePoint(point) {
        const stroke = state.draftStroke;
        if (!stroke) return;
        if (stroke.tool === "line") {
            stroke.points[1] = point;
            return;
        }

        const last = stroke.points.at(-1);
        if (Math.hypot(point.x - last.x, point.y - last.y) < 1) return;
        if (stroke.points.length >= PENCIL_POINTS_MAX) return;
        stroke.points.push(point);
    }

    function moveStroke(event) {
        if (!state.draftStroke) return;
        event.preventDefault();
        const layer = layerById(activeBoard(), state.draftStroke.layerId);
        if (!layer) return;
        appendStrokePoint(layerPoint(layer, pointerPosition(event)));
        render();
    }

    function strokeHasLength(stroke) {
        if (stroke.points.length < 2) return false;
        const first = stroke.points[0];
        const last = stroke.points.at(-1);
        return Math.hypot(last.x - first.x, last.y - first.y) >= 2;
    }

    function finishStroke(event) {
        if (!state.draftStroke) return;
        event.preventDefault();
        const stroke = state.draftStroke;
        state.draftStroke = null;
        if (!strokeHasLength(stroke)) {
            if (stroke.createdLayerId !== null) {
                const index = activeBoard().layers.findIndex(
                    (layer) => layer.id === stroke.createdLayerId,
                );
                if (index >= 0) activeBoard().layers.splice(index, 1);
                activeBoard().activeLayerId = null;
            }
            updateControls();
            render();
            return;
        }

        const board = activeBoard();
        const layer = layerById(board, stroke.layerId);
        if (!layer) {
            render();
            return;
        }
        const keys = strokeTileKeys(layer, stroke);
        const saved = captureLayerTiles(layer, keys);
        const inkOpsBefore = layer.inkOps;
        layer.inkContext.save();
        layer.inkContext.lineCap = "round";
        layer.inkContext.lineJoin = "round";
        drawStroke(layer.inkContext, stroke);
        layer.inkContext.restore();
        layer.inkOps += 1;
        addHistory(board, {
            type: "paint-layer",
            layerId: layer.id,
            createdLayerId: stroke.createdLayerId,
            tiles: saved.tiles,
            inkOpsBefore,
            bytes: saved.bytes,
        });
        setStatus(
            stroke.tool === "pencil"
                ? `Pencil stroke added to ${layer.name}.`
                : `Line added to ${layer.name}.`,
        );
        updateControls();
        render();
    }

    function cancelStroke() {
        if (!state.draftStroke) return;
        const board = activeBoard();
        const createdLayerId = state.draftStroke.createdLayerId;
        state.draftStroke = null;
        if (createdLayerId !== null) {
            const index = board.layers.findIndex(
                (layer) => layer.id === createdLayerId,
            );
            if (index >= 0 && board.layers[index].inkOps === 0) {
                board.layers.splice(index, 1);
                board.activeLayerId = null;
                updateControls();
            }
        }
        render();
    }

    function drawStroke(context, stroke) {
        const points = stroke.points;
        if (points.length === 0) return;

        context.beginPath();
        context.moveTo(points[0].x, points[0].y);
        if (stroke.tool === "pencil" && points.length > 2) {
            for (let index = 1; index < points.length - 1; index += 1) {
                const point = points[index];
                const next = points[index + 1];
                context.quadraticCurveTo(
                    point.x,
                    point.y,
                    (point.x + next.x) / 2,
                    (point.y + next.y) / 2,
                );
            }
            const last = points.at(-1);
            context.lineTo(last.x, last.y);
        } else {
            const last = points.at(-1);
            context.lineTo(last.x, last.y);
        }
        context.strokeStyle = stroke.color;
        context.lineWidth = stroke.width;
        context.stroke();
    }

    function draftStrokeOnCanvas(layer, stroke) {
        return {
            tool: stroke.tool,
            points: stroke.points.map((point) =>
                layerPointOnCanvas(layer, point),
            ),
            color: stroke.color,
            width: stroke.width * (layer.width / layer.contentWidth),
        };
    }

    function drawLayerContents(context, layer, x, y, width, height) {
        context.save();
        context.translate(
            x + (layer.flipX ? width : 0),
            y + (layer.flipY ? height : 0),
        );
        context.scale(layer.flipX ? -1 : 1, layer.flipY ? -1 : 1);
        if (layer.kind === "image")
            context.drawImage(layer.bitmap, 0, 0, width, height);
        context.drawImage(layer.ink, 0, 0, width, height);
        context.restore();
    }

    function renderBoard(context, board) {
        context.clearRect(0, 0, board.width, board.height);
        context.drawImage(board.background, 0, 0);
        for (const layer of board.layers) {
            if (!layer.visible) continue;
            drawLayerContents(
                context,
                layer,
                layer.x,
                layer.y,
                layer.width,
                layer.height,
            );
        }
    }

    function drawLayerThumbnail(thumbnail, layer) {
        const context = thumbnail.getContext("2d");
        if (!context) return;
        context.clearRect(0, 0, thumbnail.width, thumbnail.height);
        if (layer.kind === "ink") {
            for (let y = 0; y < thumbnail.height; y += 6) {
                for (let x = 0; x < thumbnail.width; x += 6) {
                    context.fillStyle =
                        (x / 6 + y / 6) % 2 === 0 ? "#fbf8ef" : "#e8e2d5";
                    context.fillRect(x, y, 6, 6);
                }
            }
        } else {
            context.fillStyle = "#ffffff";
            context.fillRect(0, 0, thumbnail.width, thumbnail.height);
        }
        const scale = Math.min(
            thumbnail.width / layer.width,
            thumbnail.height / layer.height,
        );
        const width = layer.width * scale;
        const height = layer.height * scale;
        const x = (thumbnail.width - width) / 2;
        const y = (thumbnail.height - height) / 2;
        drawLayerContents(context, layer, x, y, width, height);
    }

    function renderLayerPanel(board) {
        const scrollLeft = layerList.scrollLeft;
        const scrollTop = layerList.scrollTop;
        const focusedAction =
            layerList.contains(document.activeElement) &&
            document.activeElement.dataset.layerAction &&
            document.activeElement.dataset.layerId
                ? {
                      action: document.activeElement.dataset.layerAction,
                      id: document.activeElement.dataset.layerId,
                  }
                : null;
        const rows = document.createDocumentFragment();
        for (const layer of board.layers.toReversed()) {
            const row =
                layerRowTemplate.content.firstElementChild.cloneNode(true);
            const visibilityButton = row.querySelector(
                '[data-layer-action="visibility"]',
            );
            const selectButton = row.querySelector(
                '[data-layer-action="select"]',
            );
            const visibilityOn = row.querySelector(
                "[data-layer-visibility-on]",
            );
            const visibilityOff = row.querySelector(
                "[data-layer-visibility-off]",
            );
            const thumbnail = row.querySelector("[data-layer-thumbnail]");
            const name = row.querySelector("[data-layer-name]");
            const meta = row.querySelector("[data-layer-meta]");
            const selected = board.activeLayerId === layer.id;

            row.dataset.selected = String(selected);
            visibilityButton.dataset.layerId = String(layer.id);
            visibilityButton.setAttribute(
                "aria-pressed",
                String(layer.visible),
            );
            visibilityButton.setAttribute(
                "aria-label",
                `${layer.visible ? "Hide" : "Show"} ${layer.name}`,
            );
            visibilityButton.title =
                visibilityButton.getAttribute("aria-label");
            visibilityOn.hidden = !layer.visible;
            visibilityOff.hidden = layer.visible;
            selectButton.dataset.layerId = String(layer.id);
            selectButton.setAttribute("aria-pressed", String(selected));
            selectButton.setAttribute("aria-label", `Select ${layer.name}`);
            name.textContent = layer.name;
            meta.textContent =
                layer.kind === "image"
                    ? `${Math.round(layer.width)} × ${Math.round(layer.height)} · image`
                    : "Ink · full canvas";
            drawLayerThumbnail(thumbnail, layer);
            rows.append(row);
        }

        layerCount.textContent = String(board.layers.length).padStart(2, "0");
        mobileLayerCount.textContent = String(board.layers.length);
        if (board.layers.length === 0) {
            rows.append(
                layerEmptyTemplate.content.firstElementChild.cloneNode(true),
            );
        }
        layerList.replaceChildren(rows);
        layerList.scrollLeft = scrollLeft;
        layerList.scrollTop = scrollTop;
        if (focusedAction) {
            layerList
                .querySelector(
                    `[data-layer-action="${focusedAction.action}"][data-layer-id="${focusedAction.id}"]`,
                )
                ?.focus();
        }
    }

    function copyScope(board, image) {
        if (!image) {
            return {
                kind: "canvas",
                x: 0,
                y: 0,
                width: board.width,
                height: board.height,
            };
        }

        const x = Math.max(0, Math.round(image.x));
        const y = Math.max(0, Math.round(image.y));
        const right = Math.min(board.width, Math.round(image.x + image.width));
        const bottom = Math.min(
            board.height,
            Math.round(image.y + image.height),
        );
        return {
            kind: "image",
            x,
            y,
            width: Math.max(1, right - x),
            height: Math.max(1, bottom - y),
        };
    }

    function imageAtPoint(point) {
        const board = activeBoard();
        for (let index = board.layers.length - 1; index >= 0; index -= 1) {
            const image = board.layers[index];
            if (image.kind !== "image") continue;
            if (!image.visible) continue;
            const withinX =
                point.x >= image.x && point.x <= image.x + image.width;
            const withinY =
                point.y >= image.y && point.y <= image.y + image.height;
            if (withinX && withinY) return image;
        }
        return null;
    }

    function imageHandlePoints(image) {
        const left = image.x;
        const right = image.x + image.width;
        const top = image.y;
        const bottom = image.y + image.height;
        const centreX = left + image.width / 2;
        const centreY = top + image.height / 2;
        return [
            { name: "nw", x: left, y: top },
            { name: "n", x: centreX, y: top },
            { name: "ne", x: right, y: top },
            { name: "e", x: right, y: centreY },
            { name: "se", x: right, y: bottom },
            { name: "s", x: centreX, y: bottom },
            { name: "sw", x: left, y: bottom },
            { name: "w", x: left, y: centreY },
        ];
    }

    function displayScale() {
        const rect = canvas.getBoundingClientRect();
        return rect.width > 0 ? canvas.width / rect.width : 1;
    }

    function imageHandleAtPoint(image, point) {
        if (!image) return null;
        const hitRadius = IMAGE_HANDLE_HIT_CSS * displayScale();
        for (const handle of imageHandlePoints(image)) {
            if (Math.abs(point.x - handle.x) > hitRadius) continue;
            if (Math.abs(point.y - handle.y) > hitRadius) continue;
            return handle;
        }
        return null;
    }

    function imageBounds(image) {
        return {
            x: image.x,
            y: image.y,
            width: image.width,
            height: image.height,
        };
    }

    function clampImagePosition(image) {
        const board = activeBoard();
        image.x = Math.max(
            IMAGE_MIN_VISIBLE - image.width,
            Math.min(board.width - IMAGE_MIN_VISIBLE, image.x),
        );
        image.y = Math.max(
            IMAGE_MIN_VISIBLE - image.height,
            Math.min(board.height - IMAGE_MIN_VISIBLE, image.y),
        );
    }

    function drawImageSelection(context, image) {
        if (!image) return;
        const scale = displayScale();
        const handleSize = IMAGE_HANDLE_SIZE_CSS * scale;
        context.save();
        context.strokeStyle = "#0064e6";
        context.lineWidth = Math.max(1, 1.5 * scale);
        context.strokeRect(image.x, image.y, image.width, image.height);
        for (const handle of imageHandlePoints(image)) {
            context.fillStyle = "#ffffff";
            context.fillRect(
                handle.x - handleSize / 2,
                handle.y - handleSize / 2,
                handleSize,
                handleSize,
            );
            context.strokeRect(
                handle.x - handleSize / 2,
                handle.y - handleSize / 2,
                handleSize,
                handleSize,
            );
        }
        context.restore();
    }

    function beginImageInteraction(event) {
        if (!state.ready || state.tool !== "select" || event.button !== 0)
            return;

        event.preventDefault();
        const board = activeBoard();
        const point = pointerPosition(event);
        const chosen = targetImage();
        const handle = imageHandleAtPoint(chosen, point);
        if (handle) {
            canvas.setPointerCapture(event.pointerId);
            board.selection = null;
            state.selectionRequestId += 1;
            state.imageInteraction = {
                type: "resize",
                handle: handle.name,
                image: chosen,
                start: point,
                before: imageBounds(chosen),
            };
            render();
            return;
        }

        const image = imageAtPoint(point);
        if (!image) {
            board.activeLayerId = null;
            state.imageInteraction = null;
            updateControls();
            render();
            setStatus("No image selected; Copy uses the canvas.");
            return;
        }

        canvas.setPointerCapture(event.pointerId);
        board.activeLayerId = image.id;
        board.selection = null;
        state.selectionRequestId += 1;
        state.imageInteraction = {
            type: "move",
            image,
            start: point,
            before: imageBounds(image),
        };
        updateControls();
        render();
        setStatus(
            "Image selected. Drag to move it; drag a handle to resize it.",
        );
    }

    function resizeImage(image, interaction, point) {
        const before = interaction.before;
        const deltaX = point.x - interaction.start.x;
        const deltaY = point.y - interaction.start.y;
        let left = before.x;
        let right = before.x + before.width;
        let top = before.y;
        let bottom = before.y + before.height;

        if (interaction.handle.includes("w")) {
            left = Math.min(before.x + deltaX, right - IMAGE_MIN_SIZE);
        }
        if (interaction.handle.includes("e")) {
            right = Math.max(
                before.x + before.width + deltaX,
                left + IMAGE_MIN_SIZE,
            );
        }
        if (interaction.handle.includes("n")) {
            top = Math.min(before.y + deltaY, bottom - IMAGE_MIN_SIZE);
        }
        if (interaction.handle.includes("s")) {
            bottom = Math.max(
                before.y + before.height + deltaY,
                top + IMAGE_MIN_SIZE,
            );
        }
        if (state.selectAspectLocked) {
            const ratio = before.width / before.height;
            if (interaction.handle.length === 2) {
                const widthScale = (right - left) / before.width;
                const heightScale = (bottom - top) / before.height;
                const scale =
                    Math.abs(widthScale - 1) >= Math.abs(heightScale - 1)
                        ? widthScale
                        : heightScale;
                const minimumWidth = Math.max(
                    IMAGE_MIN_SIZE,
                    IMAGE_MIN_SIZE * ratio,
                );
                const width = Math.max(minimumWidth, before.width * scale);
                const height = width / ratio;
                if (interaction.handle.includes("w")) left = right - width;
                else right = left + width;
                if (interaction.handle.includes("n")) top = bottom - height;
                else bottom = top + height;
            } else if (
                interaction.handle === "e" ||
                interaction.handle === "w"
            ) {
                const minimumWidth = Math.max(
                    IMAGE_MIN_SIZE,
                    IMAGE_MIN_SIZE * ratio,
                );
                const width = Math.max(minimumWidth, right - left);
                const height = width / ratio;
                const centerY = before.y + before.height / 2;
                top = centerY - height / 2;
                bottom = top + height;
                if (interaction.handle === "w") left = right - width;
                else right = left + width;
            } else {
                const minimumHeight = Math.max(
                    IMAGE_MIN_SIZE,
                    IMAGE_MIN_SIZE / ratio,
                );
                const height = Math.max(minimumHeight, bottom - top);
                const width = height * ratio;
                const centerX = before.x + before.width / 2;
                left = centerX - width / 2;
                right = left + width;
                if (interaction.handle === "n") top = bottom - height;
                else bottom = top + height;
            }
        }

        image.x = left;
        image.y = top;
        image.width = right - left;
        image.height = bottom - top;
    }

    function moveImageInteraction(event) {
        if (!state.imageInteraction) return;
        event.preventDefault();
        const interaction = state.imageInteraction;
        const point = pointerPosition(event);
        if (interaction.type === "move") {
            interaction.image.x =
                interaction.before.x + point.x - interaction.start.x;
            interaction.image.y =
                interaction.before.y + point.y - interaction.start.y;
            clampImagePosition(interaction.image);
        } else {
            resizeImage(interaction.image, interaction, point);
        }
        render();
    }

    function finishImageInteraction(event) {
        if (!state.imageInteraction) return;
        event.preventDefault();
        const board = activeBoard();
        const interaction = state.imageInteraction;
        const after = imageBounds(interaction.image);
        const changed =
            after.x !== interaction.before.x ||
            after.y !== interaction.before.y ||
            after.width !== interaction.before.width ||
            after.height !== interaction.before.height;
        if (changed) {
            addHistory(board, {
                type: "transform-image",
                layerId: interaction.image.id,
                before: interaction.before,
            });
            setStatus(
                interaction.type === "move"
                    ? "Moved the selected image."
                    : "Resized the selected image.",
            );
        }
        state.imageInteraction = null;
        updateControls();
        render();
    }

    function cancelImageInteraction() {
        if (!state.imageInteraction) return;
        Object.assign(
            state.imageInteraction.image,
            state.imageInteraction.before,
        );
        state.imageInteraction = null;
        updateControls();
        render();
    }

    function updateHoverCursor(event) {
        if (!state.ready) return;
        if (state.tool !== "select") return;
        const point = pointerPosition(event);
        const handle = imageHandleAtPoint(targetImage(), point);
        if (handle) {
            canvas.style.cursor = IMAGE_HANDLE_CURSORS[handle.name];
        } else if (imageAtPoint(point)) {
            canvas.style.cursor = "grab";
        } else {
            canvas.style.cursor = "default";
        }
    }

    function render() {
        if (!state.ready) return;
        const context = canvas.getContext("2d");
        if (!context) return;

        const board = activeBoard();
        renderBoard(context, board);
        if (board.selection) {
            context.drawImage(
                board.selection.overlay,
                board.selection.bounds.minX,
                board.selection.bounds.minY,
            );
        }
        if (state.draftStroke) {
            const layer = layerById(board, state.draftStroke.layerId);
            if (layer?.visible) {
                drawStroke(
                    context,
                    draftStrokeOnCanvas(layer, state.draftStroke),
                );
            }
        }
        if (state.tool === "select") {
            const selectedImage = targetImage();
            if (selectedImage?.visible)
                drawImageSelection(context, selectedImage);
        }
    }

    function selectionHasPixel(mask, width, height, x, y) {
        if (x < 0 || x >= width) return false;
        if (y < 0 || y >= height) return false;
        return mask[y * width + x] !== 0;
    }

    function createSelectionOverlay(mask, width, height, bounds) {
        const overlayWidth = bounds.maxX - bounds.minX + 1;
        const overlayHeight = bounds.maxY - bounds.minY + 1;
        const overlay = createCanvas(overlayWidth, overlayHeight);
        const context = overlay.getContext("2d");
        if (!context) throw new Error("Selection overlay canvas unavailable");

        const image = context.createImageData(overlayWidth, overlayHeight);
        const data = image.data;
        for (let y = 0; y < overlayHeight; y += 1) {
            const globalY = bounds.minY + y;
            for (let x = 0; x < overlayWidth; x += 1) {
                const globalX = bounds.minX + x;
                if (!selectionHasPixel(mask, width, height, globalX, globalY))
                    continue;

                const boundary =
                    x === 0 ||
                    x === overlayWidth - 1 ||
                    y === 0 ||
                    y === overlayHeight - 1 ||
                    !selectionHasPixel(
                        mask,
                        width,
                        height,
                        globalX - 1,
                        globalY,
                    ) ||
                    !selectionHasPixel(
                        mask,
                        width,
                        height,
                        globalX + 1,
                        globalY,
                    ) ||
                    !selectionHasPixel(
                        mask,
                        width,
                        height,
                        globalX,
                        globalY - 1,
                    ) ||
                    !selectionHasPixel(
                        mask,
                        width,
                        height,
                        globalX,
                        globalY + 1,
                    );
                const offset = (y * overlayWidth + x) * 4;
                data[offset] = 0;
                data[offset + 1] = 100;
                data[offset + 2] = 230;
                data[offset + 3] = boundary ? 176 : 56;
            }
        }

        context.putImageData(image, 0, 0);
        return overlay;
    }

    function createSelection(board, mask, bounds, count) {
        const width = board.width;
        const height = board.height;
        if (mask.length !== width * height)
            throw new Error(
                "Selection mask dimensions do not match the raster",
            );
        const overlay = createSelectionOverlay(mask, width, height, bounds);
        board.selection = { mask, bounds, count, overlay };
    }

    function ensureSelectionWorker() {
        if (state.selectionWorker) return state.selectionWorker;
        const workerUrl = root.dataset.workerUrl;
        if (!workerUrl) {
            setStatus("Selection worker URL is missing.");
            return null;
        }

        try {
            const worker = new Worker(workerUrl);
            worker.addEventListener("message", handleSelectionResult);
            worker.addEventListener("error", () => {
                state.selectionRequestId += 1;
                worker.terminate();
                state.selectionWorker = null;
                setStatus(
                    "The selection worker could not load. Try refreshing the page.",
                );
            });
            state.selectionWorker = worker;
            return worker;
        } catch {
            setStatus("Could not start the selection worker in this browser.");
            return null;
        }
    }

    function handleSelectionResult(event) {
        const message = event.data;
        if (!message || message.requestId !== state.selectionRequestId) return;
        if (message.error) {
            setStatus(message.error);
            return;
        }

        const board = activeBoard();
        if (!board) return;
        if (!(message.mask instanceof ArrayBuffer)) {
            setStatus("The selection worker returned an invalid mask.");
            return;
        }

        const mask = new Uint8Array(message.mask);
        const width = board.width;
        const height = board.height;
        if (mask.length !== width * height) {
            setStatus("The selection worker returned a mismatched mask.");
            return;
        }
        if (!Number.isInteger(message.count) || message.count < 0) {
            setStatus("The selection worker returned an invalid pixel count.");
            return;
        }
        if (message.count === 0) {
            board.selection = null;
            updateControls();
            render();
            setStatus(
                "No matching pixels found. Try raising the tolerance or turning off Contiguous.",
            );
            return;
        }

        const bounds = message.bounds;
        const validBounds =
            bounds &&
            Number.isInteger(bounds.minX) &&
            Number.isInteger(bounds.minY) &&
            Number.isInteger(bounds.maxX) &&
            Number.isInteger(bounds.maxY) &&
            bounds.minX >= 0 &&
            bounds.minY >= 0 &&
            bounds.maxX >= bounds.minX &&
            bounds.maxY >= bounds.minY &&
            bounds.maxX < width &&
            bounds.maxY < height;
        if (!validBounds) {
            setStatus("The selection worker returned invalid bounds.");
            return;
        }

        createSelection(board, mask, bounds, message.count);
        updateControls();
        render();
        setStatus(
            `Selected ${message.count.toLocaleString()} pixels. Use Fill selection or click elsewhere with the wand.`,
        );
    }

    function selectAt(event) {
        if (!state.ready || state.tool !== "wand" || event.button !== 0) return;
        const worker = ensureSelectionWorker();
        if (!worker) return;

        event.preventDefault();
        const board = activeBoard();
        const point = canvasPoint(event);
        const width = board.width;
        const height = board.height;
        let imageData;
        try {
            const source = createCanvas(width, height);
            const sourceContext = source.getContext("2d", {
                willReadFrequently: true,
            });
            if (!sourceContext) throw new Error("2D canvas unavailable");
            renderBoard(sourceContext, board);
            imageData = sourceContext.getImageData(0, 0, width, height);
        } catch {
            setStatus("Could not read the canvas pixels for selection.");
            return;
        }

        const requestId = ++state.selectionRequestId;
        const tolerance = Math.round(
            (Math.max(0, Math.min(100, Number(toleranceInput.value))) * 255) /
                100,
        );
        setStatus("Selecting connected pixels…");
        try {
            worker.postMessage(
                {
                    requestId,
                    pixels: imageData.data.buffer,
                    width,
                    height,
                    seedX: point.x,
                    seedY: point.y,
                    tolerance,
                    contiguous: contiguousInput.checked,
                },
                [imageData.data.buffer],
            );
        } catch {
            state.selectionRequestId += 1;
            setStatus("Could not send pixels to the selection worker.");
        }
    }

    function setTool(tool) {
        if (!["select", "pencil", "line", "wand"].includes(tool)) return;
        if (state.draftStroke) cancelStroke();
        state.tool = tool;
        state.imageInteraction = null;
        updateControls();
        render();
        if (tool === "select") {
            setStatus(
                "Select selected. Click an image to move it; drag a handle to resize it.",
            );
        } else if (tool === "pencil") {
            setStatus("Pencil selected. Drag on the canvas to draw freehand.");
        } else if (tool === "line") {
            setStatus("Line selected. Drag on the canvas to draw.");
        } else {
            setStatus(
                "Magic wand selected. Click a colour region to select it.",
            );
        }
    }

    function addHistory(board, entry) {
        const bytes = entry.bytes || 0;
        if (bytes > HISTORY_BYTES_MAX) {
            board.history = [];
            board.historyBytes = 0;
            return false;
        }

        while (
            board.history.length > 0 &&
            (board.history.length >= HISTORY_ENTRY_MAX ||
                board.historyBytes + bytes > HISTORY_BYTES_MAX)
        ) {
            const removed = board.history.shift();
            board.historyBytes -= removed.bytes || 0;
        }
        board.history.push(entry);
        board.historyBytes += bytes;
        return true;
    }

    function tileKeysForBounds(width, height, bounds) {
        const columns = Math.ceil(width / TILE_SIZE);
        const rows = Math.ceil(height / TILE_SIZE);
        const minTileX = Math.max(0, Math.floor(bounds.minX / TILE_SIZE));
        const minTileY = Math.max(0, Math.floor(bounds.minY / TILE_SIZE));
        const maxTileX = Math.min(
            columns - 1,
            Math.floor(bounds.maxX / TILE_SIZE),
        );
        const maxTileY = Math.min(
            rows - 1,
            Math.floor(bounds.maxY / TILE_SIZE),
        );
        const keys = new Set();
        for (let tileY = minTileY; tileY <= maxTileY; tileY += 1) {
            for (let tileX = minTileX; tileX <= maxTileX; tileX += 1)
                keys.add(tileY * columns + tileX);
        }
        return Array.from(keys).sort((left, right) => left - right);
    }

    function selectedLayerTileKeys(board, layer) {
        const { mask, bounds } = board.selection;
        const keys = new Set();
        const columns = Math.ceil(layer.contentWidth / TILE_SIZE);
        const scaleX = layer.contentWidth / layer.width;
        const scaleY = layer.contentHeight / layer.height;
        for (let y = bounds.minY; y <= bounds.maxY; y += 1) {
            const row = y * board.width;
            const localY = Math.floor((y + 0.5 - layer.y) * scaleY);
            if (localY < 0 || localY >= layer.contentHeight) continue;
            for (let x = bounds.minX; x <= bounds.maxX; x += 1) {
                if (mask[row + x] === 0) continue;
                const localX = Math.floor((x + 0.5 - layer.x) * scaleX);
                if (localX < 0 || localX >= layer.contentWidth) continue;
                keys.add(
                    Math.floor(localY / TILE_SIZE) * columns +
                        Math.floor(localX / TILE_SIZE),
                );
            }
        }
        return Array.from(keys).sort((left, right) => left - right);
    }

    function captureLayerTiles(layer, keys) {
        const width = layer.contentWidth;
        const height = layer.contentHeight;
        const columns = Math.ceil(width / TILE_SIZE);
        const tiles = [];
        let bytes = 0;
        for (const key of keys) {
            const x = (key % columns) * TILE_SIZE;
            const y = Math.floor(key / columns) * TILE_SIZE;
            const tileWidth = Math.min(TILE_SIZE, width - x);
            const tileHeight = Math.min(TILE_SIZE, height - y);
            const imageData = layer.inkContext.getImageData(
                x,
                y,
                tileWidth,
                tileHeight,
            );
            tiles.push({ x, y, imageData });
            bytes += imageData.data.length;
        }
        return { tiles, bytes };
    }

    function strokeTileKeys(layer, stroke) {
        let minX = Infinity;
        let minY = Infinity;
        let maxX = -Infinity;
        let maxY = -Infinity;
        for (const point of stroke.points) {
            minX = Math.min(minX, point.x);
            minY = Math.min(minY, point.y);
            maxX = Math.max(maxX, point.x);
            maxY = Math.max(maxY, point.y);
        }
        const padding = stroke.width / 2 + 2;
        return tileKeysForBounds(layer.contentWidth, layer.contentHeight, {
            minX: Math.max(0, Math.floor(minX - padding)),
            minY: Math.max(0, Math.floor(minY - padding)),
            maxX: Math.min(layer.contentWidth - 1, Math.ceil(maxX + padding)),
            maxY: Math.min(layer.contentHeight - 1, Math.ceil(maxY + padding)),
        });
    }

    function selectedColour() {
        const match = /^#([0-9a-f]{6})$/i.exec(colorInput.value);
        if (!match) return null;
        const packed = Number.parseInt(match[1], 16);
        return [(packed >> 16) & 255, (packed >> 8) & 255, packed & 255];
    }

    function fillSelection() {
        const board = activeBoard();
        if (!board.selection || board.selection.count === 0) {
            setStatus("Use the magic wand to select an area first.");
            return;
        }
        const colour = selectedColour();
        if (!colour) {
            setStatus("Choose a valid fill colour.");
            return;
        }
        let layer = activeLayer();
        let createdLayerId = null;
        if (!layer) {
            layer = createInkLayer(board);
            board.layers.push(layer);
            board.activeLayerId = layer.id;
            createdLayerId = layer.id;
        }
        if (!layer.visible) {
            setStatus("Show the active layer before filling it.");
            return;
        }

        const width = layer.contentWidth;
        const height = layer.contentHeight;
        const { mask, bounds, count } = board.selection;
        let imageData;
        let saved;
        try {
            imageData = layer.inkContext.getImageData(0, 0, width, height);
            saved = captureLayerTiles(
                layer,
                selectedLayerTileKeys(board, layer),
            );
        } catch {
            setStatus("Could not read the canvas pixels for filling.");
            return;
        }

        const inkOpsBefore = layer.inkOps;
        const undoable = addHistory(board, {
            type: "paint-layer",
            layerId: layer.id,
            createdLayerId,
            tiles: saved.tiles,
            inkOpsBefore,
            bytes: saved.bytes,
        });

        const data = imageData.data;
        const scaleX = width / layer.width;
        const scaleY = height / layer.height;
        for (let y = bounds.minY; y <= bounds.maxY; y += 1) {
            const row = y * board.width;
            const localY = Math.floor((y + 0.5 - layer.y) * scaleY);
            if (localY < 0 || localY >= height) continue;
            for (let x = bounds.minX; x <= bounds.maxX; x += 1) {
                if (mask[row + x] === 0) continue;
                const localX = Math.floor((x + 0.5 - layer.x) * scaleX);
                if (localX < 0 || localX >= width) continue;
                const offset = (localY * width + localX) * 4;
                data[offset] = colour[0];
                data[offset + 1] = colour[1];
                data[offset + 2] = colour[2];
                data[offset + 3] = 255;
            }
        }

        layer.inkContext.putImageData(imageData, 0, 0);
        layer.inkOps += 1;
        updateControls();
        render();
        const undoNote = undoable
            ? ""
            : " This fill was too large to undo; history was cleared.";
        setStatus(
            `Filled ${count.toLocaleString()} selected pixels on ${layer.name}.${undoNote}`,
        );
    }

    function clearSelection() {
        const board = activeBoard();
        if (!board?.selection) return;
        board.selection = null;
        state.selectionRequestId += 1;
        updateControls();
        render();
        setStatus("Selection cleared. Choose a tool or click with the wand.");
    }

    function selectLayer(layerId) {
        const board = activeBoard();
        const layer = layerById(board, layerId);
        if (!layer) return;
        board.activeLayerId = layer.id;
        state.draftStroke = null;
        state.imageInteraction = null;
        setLayerSheetOpen(false);
        updateControls();
        render();
        setStatus(
            layer.visible
                ? layer.kind === "image"
                    ? `Selected ${layer.name}. Use Select to move it; drawing tools paint onto this layer.`
                    : `Selected ${layer.name}. Drawing tools paint onto this full-canvas layer.`
                : `Selected ${layer.name}, which is hidden. Show it to edit or copy it.`,
        );
    }

    function commitImageTransform(image, next, message) {
        const before = imageBounds(image);
        if (
            before.x === next.x &&
            before.y === next.y &&
            before.width === next.width &&
            before.height === next.height
        )
            return;
        Object.assign(image, next);
        clampImagePosition(image);
        addHistory(activeBoard(), {
            type: "transform-image",
            layerId: image.id,
            before,
        });
        updateControls();
        render();
        setStatus(message);
    }

    function applySelectGeometry(sourceInput) {
        const image = targetImage();
        if (!image) return;
        const x = Number(selectXInput.value);
        const y = Number(selectYInput.value);
        let width = Number(selectWidthInput.value);
        let height = Number(selectHeightInput.value);
        if (state.selectAspectLocked) {
            const ratio = image.width / image.height;
            if (sourceInput === selectWidthInput) {
                width = Math.max(IMAGE_MIN_SIZE, IMAGE_MIN_SIZE * ratio, width);
                height = width / ratio;
            } else if (sourceInput === selectHeightInput) {
                height = Math.max(
                    IMAGE_MIN_SIZE,
                    IMAGE_MIN_SIZE / ratio,
                    height,
                );
                width = height * ratio;
            }
        }
        if (
            !Number.isFinite(x) ||
            !Number.isFinite(y) ||
            !Number.isFinite(width) ||
            !Number.isFinite(height) ||
            width < IMAGE_MIN_SIZE ||
            height < IMAGE_MIN_SIZE
        ) {
            updateSelectGeometry(image);
            setStatus("Enter a valid position and a size of at least 24 px.");
            return;
        }
        commitImageTransform(
            image,
            {
                x: Math.round(x),
                y: Math.round(y),
                width: Math.round(width),
                height: Math.round(height),
            },
            `Updated ${image.name}'s position and size.`,
        );
    }

    function toggleSelectAspectLock() {
        if (!targetImage()) return;
        state.selectAspectLocked = !state.selectAspectLocked;
        updateControls();
        setStatus(
            state.selectAspectLocked
                ? "Aspect ratio locked. Width and height now change together."
                : "Aspect ratio unlocked.",
        );
    }

    function flipSelectedImage(axis) {
        const image = targetImage();
        if (!image) return;
        const board = activeBoard();
        const key = axis === "horizontal" ? "flipX" : "flipY";
        const before = image[key];
        image[key] = !before;
        state.imageInteraction = null;
        addHistory(board, {
            type: "flip-layer",
            layerId: image.id,
            key,
            before,
        });
        updateControls();
        render();
        setStatus(`Flipped ${image.name} ${axis}ly.`);
    }

    function resetSelectedImageAspect() {
        const image = targetImage();
        if (!image) return;
        const ratio = image.contentWidth / image.contentHeight;
        const width = Math.max(IMAGE_MIN_SIZE, Math.round(image.width));
        const height = Math.max(IMAGE_MIN_SIZE, Math.round(width / ratio));
        const centerY = image.y + image.height / 2;
        commitImageTransform(
            image,
            {
                x: image.x,
                y: Math.round(centerY - height / 2),
                width,
                height,
            },
            `Restored ${image.name}'s source aspect ratio.`,
        );
    }

    function centerSelectedImage() {
        const image = targetImage();
        if (!image) return;
        const board = activeBoard();
        commitImageTransform(
            image,
            {
                x: Math.round((board.width - image.width) / 2),
                y: Math.round((board.height - image.height) / 2),
                width: image.width,
                height: image.height,
            },
            `Centred ${image.name} on the canvas.`,
        );
    }

    function fitSelectedImage() {
        const image = targetImage();
        if (!image) return;
        const board = activeBoard();
        const scale = Math.min(
            board.width / image.width,
            board.height / image.height,
        );
        const width = Math.max(IMAGE_MIN_SIZE, Math.round(image.width * scale));
        const height = Math.max(
            IMAGE_MIN_SIZE,
            Math.round(image.height * scale),
        );
        commitImageTransform(
            image,
            {
                x: Math.round((board.width - width) / 2),
                y: Math.round((board.height - height) / 2),
                width,
                height,
            },
            `Fit ${image.name} to the canvas.`,
        );
    }

    function toggleLayerVisibility(layerId) {
        const board = activeBoard();
        const layer = layerById(board, layerId);
        if (!layer) return;
        if (state.draftStroke?.layerId === layer.id) cancelStroke();
        const before = layer.visible;
        layer.visible = !before;
        state.imageInteraction = null;
        addHistory(board, {
            type: "layer-visibility",
            layerId: layer.id,
            before,
        });
        updateControls();
        render();
        setStatus(`${layer.visible ? "Showed" : "Hid"} ${layer.name}.`);
    }

    function addInkLayer() {
        const board = activeBoard();
        const layer = createInkLayer(board);
        board.layers.push(layer);
        board.activeLayerId = layer.id;
        addHistory(board, {
            type: "add-layer",
            layer,
            bytes: board.width * board.height * 4,
        });
        updateControls();
        render();
        setStatus(`${layer.name} added above the current layer.`);
    }

    function moveLayer(direction) {
        const board = activeBoard();
        const layer = activeLayer();
        if (!layer) return;
        const fromIndex = board.layers.indexOf(layer);
        const toIndex = fromIndex + direction;
        if (toIndex < 0) return;
        if (toIndex >= board.layers.length) return;
        board.layers.splice(fromIndex, 1);
        board.layers.splice(toIndex, 0, layer);
        addHistory(board, {
            type: "move-layer",
            layerId: layer.id,
            fromIndex,
        });
        updateControls();
        render();
        setStatus(
            direction > 0
                ? `Brought ${layer.name} forward.`
                : `Sent ${layer.name} backward.`,
        );
    }

    async function duplicateLayer() {
        const board = activeBoard();
        const source = activeLayer();
        if (!source) return;
        let duplicate;
        if (source.kind === "image") {
            let bitmap;
            try {
                bitmap = await createImageBitmap(source.bitmap);
            } catch {
                setStatus("Could not duplicate that image layer.");
                return;
            }
            duplicate = createImageLayer(board, bitmap, {
                x: source.x + PASTE_OFFSET_PIXELS,
                y: source.y + PASTE_OFFSET_PIXELS,
                width: source.width,
                height: source.height,
            });
            clampImagePosition(duplicate);
        } else {
            duplicate = createInkLayer(board);
        }
        if (activeBoard() !== board || !board.layers.includes(source)) {
            if (duplicate.bitmap) duplicate.bitmap.close();
            return;
        }

        duplicate.inkContext.drawImage(source.ink, 0, 0);
        duplicate.inkOps = source.inkOps;
        duplicate.visible = source.visible;
        duplicate.flipX = source.flipX;
        duplicate.flipY = source.flipY;
        board.layers.splice(board.layers.indexOf(source) + 1, 0, duplicate);
        board.activeLayerId = duplicate.id;
        addHistory(board, {
            type: "duplicate-layer",
            layer: duplicate,
            sourceLayerId: source.id,
            bytes:
                duplicate.contentWidth * duplicate.contentHeight * 4 +
                (duplicate.bitmap
                    ? duplicate.bitmap.width * duplicate.bitmap.height * 4
                    : 0),
        });
        updateControls();
        render();
        setStatus(`Duplicated ${source.name} as ${duplicate.name}.`);
    }

    function deleteActiveLayer() {
        const board = activeBoard();
        const layer = activeLayer();
        if (!layer) return;
        const index = board.layers.indexOf(layer);
        if (index < 0) return;
        addHistory(board, { type: "delete-layer", layer, index });
        board.layers.splice(index, 1);
        board.activeLayerId =
            board.layers[Math.min(index, board.layers.length - 1)]?.id ?? null;
        state.imageInteraction = null;
        updateControls();
        render();
        setStatus(`Deleted ${layer.name}.`);
    }

    function undo() {
        const board = activeBoard();
        if (!board) return;
        const entry = board.history.pop();
        if (!entry) return;
        board.historyBytes -= entry.bytes || 0;

        if (entry.type === "paint-layer") {
            const layer = layerById(board, entry.layerId);
            if (layer) {
                for (const tile of entry.tiles) {
                    layer.inkContext.putImageData(
                        tile.imageData,
                        tile.x,
                        tile.y,
                    );
                }
                layer.inkOps = entry.inkOpsBefore;
                if (entry.createdLayerId !== null) {
                    const index = board.layers.findIndex(
                        (value) => value.id === entry.createdLayerId,
                    );
                    if (index >= 0) board.layers.splice(index, 1);
                    board.activeLayerId = null;
                }
            }
            setStatus("Undid the last paint stroke or fill.");
        } else if (entry.type === "clear-layer-ink") {
            const layer = layerById(board, entry.layerId);
            if (layer) {
                layer.inkContext.putImageData(entry.imageData, 0, 0);
                layer.inkOps = entry.inkOpsBefore;
            }
            setStatus("Restored the cleared ink.");
        } else if (entry.type === "add-images") {
            const ids = new Set(entry.layers.map((layer) => layer.id));
            for (const layer of entry.layers) layer.bitmap.close();
            board.layers = board.layers.filter((layer) => !ids.has(layer.id));
            board.activeLayerId = null;
            setStatus("Undid the last paste.");
        } else if (entry.type === "transform-image") {
            const image = layerById(board, entry.layerId);
            if (image) Object.assign(image, entry.before);
            setStatus("Undid the image transform.");
        } else if (entry.type === "flip-layer") {
            const layer = layerById(board, entry.layerId);
            if (layer) layer[entry.key] = entry.before;
            setStatus("Undid the image flip.");
        } else if (entry.type === "layer-visibility") {
            const layer = layerById(board, entry.layerId);
            if (layer) layer.visible = entry.before;
            setStatus("Restored the layer visibility.");
        } else if (entry.type === "move-layer") {
            const layer = layerById(board, entry.layerId);
            if (layer) {
                board.layers.splice(board.layers.indexOf(layer), 1);
                board.layers.splice(entry.fromIndex, 0, layer);
            }
            setStatus("Restored the layer order.");
        } else if (entry.type === "duplicate-layer") {
            const index = board.layers.findIndex(
                (layer) => layer.id === entry.layer.id,
            );
            if (index >= 0) board.layers.splice(index, 1);
            if (board.activeLayerId === entry.layer.id) {
                board.activeLayerId = board.layers.some(
                    (layer) => layer.id === entry.sourceLayerId,
                )
                    ? entry.sourceLayerId
                    : null;
            }
            if (entry.layer.bitmap) entry.layer.bitmap.close();
            setStatus("Removed the duplicated layer.");
        } else if (entry.type === "delete-layer") {
            board.layers.splice(
                Math.min(entry.index, board.layers.length),
                0,
                entry.layer,
            );
            board.activeLayerId = entry.layer.id;
            setStatus(`Restored ${entry.layer.name}.`);
        } else if (entry.type === "add-layer") {
            const index = board.layers.findIndex(
                (layer) => layer.id === entry.layer.id,
            );
            if (index >= 0) board.layers.splice(index, 1);
            if (board.activeLayerId === entry.layer.id)
                board.activeLayerId = null;
            setStatus("Removed the new layer.");
        }

        board.selection = null;
        state.selectionRequestId += 1;
        updateControls();
        render();
    }

    function clearDrawing() {
        const board = activeBoard();
        const layer = activeLayer();
        if (!layer || layer.inkOps === 0) return;
        let imageData;
        try {
            imageData = layer.inkContext.getImageData(
                0,
                0,
                layer.contentWidth,
                layer.contentHeight,
            );
        } catch {
            setStatus("Could not read this layer's ink.");
            return;
        }
        addHistory(board, {
            type: "clear-layer-ink",
            layerId: layer.id,
            imageData,
            inkOpsBefore: layer.inkOps,
            bytes: imageData.data.length,
        });
        layer.inkContext.clearRect(
            0,
            0,
            layer.contentWidth,
            layer.contentHeight,
        );
        layer.inkOps = 0;
        updateControls();
        render();
        setStatus(`Ink cleared from ${layer.name}.`);
    }

    function imageFiles(list) {
        return Array.from(list || []).filter(
            (file) => file instanceof Blob && file.type.startsWith("image/"),
        );
    }

    function clipboardImageFiles(event) {
        const itemFiles = imageFiles(
            Array.from(event.clipboardData?.items || []).map((item) =>
                item.getAsFile(),
            ),
        );
        return itemFiles.length > 0
            ? itemFiles
            : imageFiles(event.clipboardData?.files);
    }

    async function pasteFromClipboard() {
        if (!navigator.clipboard?.read) {
            setStatus(
                "Clipboard access is unavailable. Use Ctrl/Cmd+V instead.",
            );
            return;
        }
        try {
            const items = await navigator.clipboard.read();
            const images = [];
            for (const item of items) {
                const type = item.types.find((value) =>
                    value.startsWith("image/"),
                );
                if (type) images.push(await item.getType(type));
            }
            if (images.length > 0) {
                await loadImages(images);
            } else {
                setStatus("No images found on the clipboard.");
            }
        } catch {
            setStatus("Clipboard access was blocked. Use Ctrl/Cmd+V instead.");
        }
    }

    async function copyImage() {
        const board = activeBoard();
        if (!board) return;
        const selectedImage = targetImage();
        const scope = copyScope(
            board,
            selectedImage?.visible ? selectedImage : null,
        );
        if (!navigator.clipboard?.write || !window.ClipboardItem) {
            setStatus("Copying images is unavailable in this browser.");
            return;
        }

        setStatus(`Preparing ${scope.kind} PNG…`);
        try {
            const source = createCanvas(board.width, board.height);
            const sourceContext = source.getContext("2d");
            if (!sourceContext) throw new Error("2D canvas unavailable");
            renderBoard(sourceContext, board);

            const output = document.createElement("canvas");
            output.width = scope.width;
            output.height = scope.height;
            const context = output.getContext("2d");
            if (!context) throw new Error("2D canvas unavailable");
            context.drawImage(
                source,
                scope.x,
                scope.y,
                scope.width,
                scope.height,
                0,
                0,
                scope.width,
                scope.height,
            );
            const blob = await new Promise((resolve) =>
                output.toBlob(resolve, "image/png"),
            );
            if (!blob) throw new Error("PNG encoding failed");
            await navigator.clipboard.write([
                new ClipboardItem({ "image/png": blob }),
            ]);
            setStatus(
                `Copied ${scope.kind} · ${scope.width} × ${scope.height}.`,
            );
        } catch {
            setStatus(
                "Could not copy. Check clipboard permissions and try again.",
            );
        }
    }

    async function initializeEditor() {
        try {
            state.board = blankBoard();
            const board = activeBoard();
            canvas.width = board.width;
            canvas.height = board.height;
            dropZone.style.aspectRatio = `${board.width} / ${board.height}`;
            dropZone.style.setProperty(
                "--canvas-aspect",
                String(board.width / board.height),
            );
            state.ready = true;
            updateControls();
            render();
            setStatus(blankStatus(board));
        } catch {
            setStatus(
                "This browser could not initialize a 2D canvas for the editor.",
            );
        }
    }

    window.addEventListener("paste", (event) => {
        const images = clipboardImageFiles(event);
        if (images.length === 0) return;
        event.preventDefault();
        void loadImages(images);
    });

    pasteButton.addEventListener("click", () => void pasteFromClipboard());
    fileInput.addEventListener("change", () => {
        const images = imageFiles(fileInput.files);
        fileInput.value = "";
        if (images.length > 0) void loadImages(images);
    });
    canvas.addEventListener("pointerdown", (event) => {
        state.pastePoint = canvasPoint(event);
        updateCursorReadout(event);
        if (state.tool === "wand") {
            selectAt(event);
            return;
        }
        if (state.tool === "select") {
            beginImageInteraction(event);
            return;
        }
        beginStroke(event);
    });
    canvas.addEventListener("pointermove", (event) => {
        state.pastePoint = canvasPoint(event);
        updateCursorReadout(event);
        if (state.imageInteraction) {
            moveImageInteraction(event);
        } else if (state.draftStroke) {
            moveStroke(event);
        } else {
            updateHoverCursor(event);
        }
    });
    canvas.addEventListener("pointerleave", () => {
        state.pastePoint = null;
        cursorReadout.textContent = "---- / ----";
    });
    canvas.addEventListener("pointerup", (event) => {
        if (state.draftStroke) {
            finishStroke(event);
        } else {
            finishImageInteraction(event);
        }
    });
    canvas.addEventListener("pointercancel", () => {
        cancelStroke();
        cancelImageInteraction();
    });
    undoButton.addEventListener("click", undo);
    clearButton.addEventListener("click", clearDrawing);
    copyButton.addEventListener("click", () => void copyImage());
    mobileAddButton.addEventListener("click", () => fileInput.click());
    emptyAddButton.addEventListener("click", () => fileInput.click());
    mobileCopyButton.addEventListener("click", () => void copyImage());
    mobileLayersButton.addEventListener("click", () => {
        setLayerSheetOpen(layerPanel.dataset.sheetOpen !== "true");
    });
    layerBackdrop.addEventListener("click", () => setLayerSheetOpen(false));
    layerCloseButton.addEventListener("click", () => setLayerSheetOpen(false));
    fillButton.addEventListener("click", fillSelection);
    clearSelectionButton.addEventListener("click", clearSelection);
    layerList.addEventListener("click", (event) => {
        const target = event.target;
        if (!(target instanceof Element)) return;
        const button = target.closest("button[data-layer-action]");
        if (!button || !layerList.contains(button)) return;
        const layerId = Number(button.dataset.layerId);
        if (!Number.isInteger(layerId)) return;
        if (button.dataset.layerAction === "select") {
            selectLayer(layerId);
        } else if (button.dataset.layerAction === "visibility") {
            toggleLayerVisibility(layerId);
        }
    });
    layerNewButton.addEventListener("click", addInkLayer);
    layerBackwardButton.addEventListener("click", () => moveLayer(-1));
    layerForwardButton.addEventListener("click", () => moveLayer(1));
    layerDuplicateButton.addEventListener("click", () => void duplicateLayer());
    layerDeleteButton.addEventListener("click", deleteActiveLayer);
    selectAspectButton.addEventListener("click", toggleSelectAspectLock);
    selectCenterButton.addEventListener("click", centerSelectedImage);
    selectFitButton.addEventListener("click", fitSelectedImage);
    selectFlipHorizontalButton.addEventListener("click", () =>
        flipSelectedImage("horizontal"),
    );
    selectFlipVerticalButton.addEventListener("click", () =>
        flipSelectedImage("vertical"),
    );
    selectResetAspectButton.addEventListener("click", resetSelectedImageAspect);
    for (const input of [
        selectXInput,
        selectYInput,
        selectWidthInput,
        selectHeightInput,
    ]) {
        input.addEventListener("change", () => applySelectGeometry(input));
        input.addEventListener("keydown", (event) => {
            if (event.key !== "Enter") return;
            event.preventDefault();
            applySelectGeometry(input);
            input.blur();
        });
    }
    widthInput.addEventListener("input", updateWidthLabel);
    toleranceInput.addEventListener("input", updateToleranceLabel);
    colorToggle.addEventListener("click", () => {
        setColorPickerOpen(colorPopover.hidden);
    });
    colorDone.addEventListener("click", () => {
        setColorPickerOpen(false);
        colorToggle.focus();
    });
    colorPlane.addEventListener("pointerdown", (event) => {
        if (event.button !== 0) return;
        event.preventDefault();
        colorPlanePointerId = event.pointerId;
        colorPlane.setPointerCapture(event.pointerId);
        updatePlaneFromPointer(event);
    });
    colorPlane.addEventListener("pointermove", (event) => {
        if (event.pointerId !== colorPlanePointerId) return;
        updatePlaneFromPointer(event);
    });
    colorPlane.addEventListener("pointerup", (event) => {
        if (event.pointerId !== colorPlanePointerId) return;
        colorPlanePointerId = null;
        if (colorPlane.hasPointerCapture(event.pointerId)) {
            colorPlane.releasePointerCapture(event.pointerId);
        }
    });
    colorPlane.addEventListener("pointercancel", (event) => {
        if (event.pointerId !== colorPlanePointerId) return;
        colorPlanePointerId = null;
    });
    colorPlane.addEventListener("keydown", (event) => {
        if (!pickerHsv) return;
        const step = event.shiftKey ? 0.05 : 0.01;
        let saturation = pickerHsv.s;
        let value = pickerHsv.v;
        if (event.key === "ArrowLeft") saturation -= step;
        else if (event.key === "ArrowRight") saturation += step;
        else if (event.key === "ArrowUp") value += step;
        else if (event.key === "ArrowDown") value -= step;
        else return;
        event.preventDefault();
        const hsv = {
            h: pickerHsv.h,
            s: clampUnit(saturation),
            v: clampUnit(value),
        };
        setInkColor(hsvToHex(hsv), hsv);
    });
    colorHue.addEventListener("pointerdown", (event) => {
        if (event.button !== 0) return;
        event.preventDefault();
        colorHuePointerId = event.pointerId;
        colorHue.setPointerCapture(event.pointerId);
        updateHueFromPointer(event);
    });
    colorHue.addEventListener("pointermove", (event) => {
        if (event.pointerId !== colorHuePointerId) return;
        updateHueFromPointer(event);
    });
    colorHue.addEventListener("pointerup", (event) => {
        if (event.pointerId !== colorHuePointerId) return;
        colorHuePointerId = null;
        if (colorHue.hasPointerCapture(event.pointerId)) {
            colorHue.releasePointerCapture(event.pointerId);
        }
    });
    colorHue.addEventListener("pointercancel", (event) => {
        if (event.pointerId !== colorHuePointerId) return;
        colorHuePointerId = null;
    });
    colorHue.addEventListener("keydown", (event) => {
        if (!pickerHsv) return;
        const step = event.shiftKey ? 10 : 1;
        let hue = pickerHsv.h;
        if (event.key === "ArrowLeft" || event.key === "ArrowDown") {
            hue = (hue - step + 360) % 360;
        } else if (event.key === "ArrowRight" || event.key === "ArrowUp") {
            hue = (hue + step) % 360;
        } else if (event.key === "Home") {
            hue = 0;
        } else if (event.key === "End") {
            hue = 359.999;
        } else {
            return;
        }
        event.preventDefault();
        const hsv = { ...pickerHsv, h: hue };
        setInkColor(hsvToHex(hsv), hsv);
    });
    colorHex.addEventListener("input", () => {
        if (normalizeHex(colorHex.value)) {
            applyHexInput();
        } else {
            colorHex.setAttribute("aria-invalid", "true");
        }
    });
    colorHex.addEventListener("blur", restoreHexInput);
    colorHex.addEventListener("keydown", (event) => {
        if (event.key !== "Enter") return;
        event.preventDefault();
        applyHexInput();
        colorHex.select();
    });
    for (const preset of colorPresets) {
        preset.addEventListener("click", () =>
            setInkColor(preset.dataset.editorColor),
        );
    }
    for (const button of toolButtons) {
        button.addEventListener("click", () =>
            setTool(button.dataset.editorTool),
        );
    }

    document.addEventListener("pointerdown", (event) => {
        if (colorPopover.hidden) return;
        if (
            event.target instanceof Node &&
            colorControl.contains(event.target)
        ) {
            return;
        }
        setColorPickerOpen(false);
    });

    dropZone.addEventListener("dragover", (event) => {
        event.preventDefault();
        dropZone.style.borderColor = "#0064e6";
        setStatus("Drop the images to paste them onto the canvas.");
    });
    dropZone.addEventListener("dragleave", () => {
        dropZone.style.borderColor = "";
    });
    dropZone.addEventListener("drop", (event) => {
        event.preventDefault();
        dropZone.style.borderColor = "";
        state.pastePoint = canvasPoint(event);
        const images = imageFiles(event.dataTransfer?.files);
        if (images.length > 0) {
            void loadImages(images);
        } else {
            setStatus("Drop image files to paste them.");
        }
    });

    window.addEventListener("keydown", (event) => {
        if (event.key === "Escape" && !colorPopover.hidden) {
            event.preventDefault();
            setColorPickerOpen(false);
            colorToggle.focus();
            return;
        }
        const target = event.target;
        const editing =
            target instanceof HTMLElement &&
            (target.isContentEditable ||
                target instanceof HTMLInputElement ||
                target instanceof HTMLTextAreaElement ||
                target instanceof HTMLSelectElement);
        if (editing) return;

        if (event.key === "Escape" && layerPanel.dataset.sheetOpen === "true") {
            event.preventDefault();
            setLayerSheetOpen(false);
            return;
        }

        if (event.key === "Escape") {
            const board = activeBoard();
            if (board?.selection) {
                event.preventDefault();
                clearSelection();
            } else if (board?.activeLayerId !== null) {
                event.preventDefault();
                board.activeLayerId = null;
                updateControls();
                render();
                setStatus("Layer deselected; Copy uses the canvas.");
            }
            return;
        }

        if (
            (event.key === "Delete" || event.key === "Backspace") &&
            activeLayer()
        ) {
            event.preventDefault();
            deleteActiveLayer();
            return;
        }

        const modifier = event.metaKey || event.ctrlKey;
        if (!modifier) return;
        if (event.key.toLowerCase() === "z" && !event.shiftKey) {
            event.preventDefault();
            undo();
        }
        if (
            event.key.toLowerCase() === "c" &&
            !window.getSelection()?.toString()
        ) {
            event.preventDefault();
            void copyImage();
        }
        if (event.key.toLowerCase() === "d" && !event.shiftKey) {
            event.preventDefault();
            void duplicateLayer();
        }
    });

    updateWidthLabel();
    updateToleranceLabel();
    setInkColor(colorInput.value);
    initialization = initializeEditor();
    void initialization;
})();
