import { z } from "zod";

// 1. Schema thuần để Validate Form (Giữ nguyên cấu trúc UI cần)
export const ClipConfigSchema = z.object({
    prompt: z.string(),
    clipCount: z.number().min(1).max(10),
    aspectRatio: z.enum(["9:16", "16:9", "1:1", "3:4"]),
    subtitle: z.boolean(),
});

// 2. Tạo bảng quy đổi
const DIMENSIONS = {
    "9:16": { width: 1080, height: 1920 },
    "3:4": { width: 1080, height: 1440 },
    "16:9": { width: 1920, height: 1080 },
    "1:1": { width: 1080, height: 1080 },
};

// 3. Hàm hỗ trợ Transform dữ liệu (Sẽ gọi hàm này trong handleUpload)
export const transformToApiData = (data: z.infer<typeof ClipConfigSchema>) => {
    const ratio = data.aspectRatio as keyof typeof DIMENSIONS;
    const { width, height } = DIMENSIONS[ratio];

    return {
        prompt: data.prompt,
        clipCount: data.clipCount,
        subtitle: data.subtitle,
        targetWidth: width,
        targetHeight: height,
    };
};