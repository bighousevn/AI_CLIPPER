export interface ClipConfig {
    prompt: string;
    clip_count: number;
    aspect_ratio: "9:16" | "16:9" | "1:1" | "3:4"  // sửa đây
    subtitle: boolean;
}

export interface ClipConfigAPI {
    prompt: string;
    clip_count: number;
    target_width: number;
    target_height: number;
    subtitle: boolean
}