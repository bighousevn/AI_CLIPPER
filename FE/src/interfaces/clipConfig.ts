export interface ClipConfig {
    prompt: string;
    clipCount: number;
    aspectRatio: "9:16" | "16:9" | "1:1"; // sửa đây
    subtitle: boolean;
}
