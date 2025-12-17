export interface ClipConfig {
    prompt: string;
    clipCount: number;
    aspectRatio: "9:16" | "16:9" | "1:1" | "3:4"  // sửa đây
    subtitle: boolean;
}

export interface ClipConfigAPI {
    prompt: string;
    clipCount: number;
    targetWidth: number;
    targetHeight: number;
    subtitle: boolean
}