import glob
import json
import os
import pathlib
import pickle
import shutil
import subprocess
import time
import uuid
import cv2
from fastapi import Depends, HTTPException, status
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer
import ffmpegcv
import numpy as np
import pysubs2
from tqdm import tqdm
import whisperx
from google import genai
import google.genai.errors
import modal
from typing import List, Dict, Tuple, Optional  

from pydantic import BaseModel

class VideoConfig(BaseModel):
    prompt: str
    clip_count: int
    target_width: int
    target_height: int
    subtitle: bool

class ProcessVideoRequest(BaseModel):
    storage_path: str
    config : VideoConfig
    


image = (modal.Image.from_registry(
    "nvidia/cuda:12.8.0-devel-ubuntu22.04", add_python="3.12")
    .apt_install(["ffmpeg", "libgl1-mesa-glx", "wget", "libcudnn8", "libcudnn8-dev"])
    .pip_install_from_requirements("requirements.txt")
    # .pip_install("supabase",force_build=True)
    .run_commands(["mkdir -p /usr/share/fonts/truetype/custom",
                   "wget -O /usr/share/fonts/truetype/custom/Anton-Regular.ttf https://github.com/google/fonts/raw/main/ofl/anton/Anton-Regular.ttf",
                   "fc-cache -f -v",
                   "echo '--- Installed pip packages START ---'", 
        "pip list",                                 
        "echo '--- Installed pip packages END ---'"])
    .add_local_dir("asd", "/asd", copy=True))

app = modal.App("ai-podcast-clipper", image=image)

volume = modal.Volume.from_name(
    "ai-podcast-clipper-model-cache", create_if_missing=True
)

mount_path = "/root/.cache/torch"

auth_scheme = HTTPBearer()

def create_vertical_video(tracks, scores, pyframes_path, pyavi_path, audio_path, output_path, framerate=25, target_width=1080, target_height=1920):

    flist = glob.glob(os.path.join(pyframes_path, "*.jpg"))
    flist.sort()

    faces = [[] for _ in range(len(flist))]

    for tidx, track in enumerate(tracks):
        score_array = scores[tidx]
        for fidx, frame in enumerate(track["track"]["frame"].tolist()):
            slice_start = max(fidx - 30, 0)
            slice_end = min(fidx + 30, len(score_array))
            score_slice = score_array[slice_start:slice_end]
            avg_score = float(np.mean(score_slice)
                              if len(score_slice) > 0 else 0)

            faces[frame].append(
                {'track': tidx, 'score': avg_score, 's': track['proc_track']["s"][fidx], 'x': track['proc_track']["x"][fidx], 'y': track['proc_track']["y"][fidx]})

    temp_video_path = os.path.join(pyavi_path, "video_only.mp4")
 

    vout = None
    for fidx, fname in tqdm(enumerate(flist), total=len(flist), desc="Creating vertical video"):
        img = cv2.imread(fname)
        if img is None:
            continue

        current_faces = faces[fidx]

        max_score_face = max(
            current_faces, key=lambda face: face['score']) if current_faces else None

        if max_score_face and max_score_face['score'] < 0:
            max_score_face = None

        if vout is None:
            vout = ffmpegcv.VideoWriterNV(
                file=temp_video_path,
                codec=None,
                fps=framerate,
                resize=(target_width, target_height)
            )

        if max_score_face:
            mode = "crop"
        else:
            mode = "resize"

        if mode == "resize":
            scale = target_width / img.shape[1]
            resized_height = int(img.shape[0] * scale)
            resized_image = cv2.resize(
                img, (target_width, resized_height), interpolation=cv2.INTER_AREA)

            scale_for_bg = max(
                target_width / img.shape[1], target_height / img.shape[0])
            bg_width = int(img.shape[1] * scale_for_bg)
            bg_heigth = int(img.shape[0] * scale_for_bg)

            blurred_background = cv2.resize(img, (bg_width, bg_heigth))
            blurred_background = cv2.GaussianBlur(
                blurred_background, (121, 121), 0)

            crop_x = (bg_width - target_width) // 2
            crop_y = (bg_heigth - target_height) // 2
            blurred_background = blurred_background[crop_y:crop_y +
                                                    target_height, crop_x:crop_x + target_width]

            center_y = (target_height - resized_height) // 2
            blurred_background[center_y:center_y +
                               resized_height, :] = resized_image

            vout.write(blurred_background)

        elif mode == "crop":
            scale = target_height / img.shape[0]
            resized_image = cv2.resize(
                img, None, fx=scale, fy=scale, interpolation=cv2.INTER_AREA)
            frame_width = resized_image.shape[1]

            center_x = int(
                max_score_face["x"] * scale if max_score_face else frame_width // 2)
            top_x = max(min(center_x - target_width // 2,
                        frame_width - target_width), 0)

            image_cropped = resized_image[0:target_height,
                                          top_x:top_x + target_width]

            vout.write(image_cropped)

    if vout:
        vout.release()

    ffmpeg_command = (f"ffmpeg -y -i {temp_video_path} -i {audio_path} "
                      f"-c:v h264 -preset fast -crf 23 -c:a aac -b:a 128k "
                      f"{output_path}")
    subprocess.run(ffmpeg_command, shell=True, check=True, text=True)


# def create_subtitles_with_ffmpeg(transcript_segments: list, clip_start: float, clip_end: float, clip_video_path: str, output_path: str, max_words: int = 5):
#     temp_dir = os.path.dirname(output_path)
#     subtitle_path = os.path.join(temp_dir, "temp_subtitles.ass")

#     clip_segments = [segment for segment in transcript_segments
#                      if segment.get("start") is not None
#                      and segment.get("end") is not None
#                      and segment.get("end") > clip_start
#                      and segment.get("start") < clip_end
#                      ]

#     subtitles = []
#     current_words = []
#     current_start = None
#     current_end = None

#     for segment in clip_segments:
#         word = segment.get("word", "").strip()
#         seg_start = segment.get("start")
#         seg_end = segment.get("end")

#         if not word or seg_start is None or seg_end is None:
#             continue

#         start_rel = max(0.0, seg_start - clip_start)
#         end_rel = max(0.0, seg_end - clip_start)

#         if end_rel <= 0:
#             continue

#         if not current_words:
#             current_start = start_rel
#             current_end = end_rel
#             current_words = [word]
#         elif len(current_words) >= max_words:
#             subtitles.append(
#                 (current_start, current_end, ' '.join(current_words)))
#             current_words = [word]
#             current_start = start_rel
#             current_end = end_rel
#         else:
#             current_words.append(word)
#             current_end = end_rel

#     if current_words:
#         subtitles.append(
#             (current_start, current_end, ' '.join(current_words)))

#     subs = pysubs2.SSAFile()

#     subs.info["WrapStyle"] = 0
#     subs.info["ScaledBorderAndShadow"] = "yes"
#     subs.info["PlayResX"] = 1080
#     subs.info["PlayResY"] = 1920
#     subs.info["ScriptType"] = "v4.00+"

#     style_name = "Default"
#     new_style = pysubs2.SSAStyle()
#     new_style.fontname = "Anton"
#     new_style.fontsize = 140
#     new_style.primarycolor = pysubs2.Color(255, 255, 255)
#     new_style.outline = 2.0
#     new_style.shadow = 2.0
#     new_style.shadowcolor = pysubs2.Color(0, 0, 0, 128)
#     new_style.alignment = 2
#     new_style.marginl = 50
#     new_style.marginr = 50
#     new_style.marginv = 50
#     new_style.spacing = 0.0

#     subs.styles[style_name] = new_style

#     for i, (start, end, text) in enumerate(subtitles):
#         start_time = pysubs2.make_time(s=start)
#         end_time = pysubs2.make_time(s=end)
#         line = pysubs2.SSAEvent(
#             start=start_time, end=end_time, text=text, style=style_name)
#         subs.events.append(line)

#     subs.save(subtitle_path)

#     ffmpeg_cmd = (f"ffmpeg -y -i {clip_video_path} -vf \"ass={subtitle_path}\" "
#                   f"-c:v h264 -preset fast -crf 23 {output_path}")

#     subprocess.run(ffmpeg_cmd, shell=True, check=True)


def _build_karaoke_line(words_info: List[Dict]) -> Optional[Tuple[float, float, str, List[Dict]]]:
    """
    Trả về (line_start_rel_s, line_end_rel_s, full_text, words_info)
    words_info: list dict có keys: word, start_rel, end_rel (relative to clip_start, seconds)
    """
    if not words_info:
        return None
    line_start = words_info[0]["start_rel"]
    line_end = words_info[-1]["end_rel"]
    full_text = " ".join(w["word"] for w in words_info)
    return (line_start, line_end, full_text, words_info)


def create_subtitles_with_ffmpeg(
    transcript_segments: list,
    clip_start: float,
    clip_end: float,
    clip_video_path: str,
    output_path: str,
    max_words: int = 5
) -> None:
  

    temp_dir = os.path.dirname(output_path) or "."
    os.makedirs(temp_dir, exist_ok=True)
    subtitle_path = os.path.join(temp_dir, "temp_subtitles.ass")

    # Lọc segments nằm trong clip
    clip_segments = [
        segment for segment in transcript_segments
        if segment.get("start") is not None
           and segment.get("end") is not None
           and segment.get("end") > clip_start
           and segment.get("start") < clip_end
    ]

    # Gom thành dòng theo max_words
    subtitles_lines: List[List[Dict]] = []
    current_line_words_info: List[Dict] = []

    for segment in clip_segments:
        word = (segment.get("word") or "").strip()
        seg_start = segment.get("start")
        seg_end = segment.get("end")

        if not word or seg_start is None or seg_end is None:
            continue

        # relative thời gian (so với clip_start)
        start_rel = max(0.0, seg_start - clip_start)
        end_rel = max(0.0, seg_end - clip_start)

        if end_rel <= 0:
            continue

        word_info = {
            "word": word,
            "start": seg_start,
            "end": seg_end,
            "start_rel": start_rel,
            "end_rel": end_rel
        }

        if not current_line_words_info:
            current_line_words_info.append(word_info)
        elif len(current_line_words_info) >= max_words:
            subtitles_lines.append(current_line_words_info)
            current_line_words_info = [word_info]
        else:
            current_line_words_info.append(word_info)

    if current_line_words_info:
        subtitles_lines.append(current_line_words_info)

    # --- Tạo file ASS với 1 style karaoke duy nhất ---
    subs = pysubs2.SSAFile()

    subs.info["WrapStyle"] = 0
    subs.info["ScaledBorderAndShadow"] = "yes"
    subs.info["PlayResX"] = 1080
    subs.info["PlayResY"] = 1920
    subs.info["ScriptType"] = "v4.00+"

    # Style karaoke: primary = highlight (vàng), secondary = un-highlight (trắng)
    k_style = pysubs2.SSAStyle()
    k_style.fontname = "Anton"
    k_style.fontsize = 140
    k_style.primarycolor = pysubs2.Color(255, 255, 0)      # màu vàng (highlight)
    k_style.secondarycolor = pysubs2.Color(255, 255, 255)  # màu trắng (chưa đến lượt)
    k_style.outline = 2.0
    k_style.shadow = 2.0
    k_style.shadowcolor = pysubs2.Color(0, 0, 0, 128)
    k_style.alignment = 2        # bottom-center
    k_style.marginl = 50
    k_style.marginr = 50
    k_style.marginv = 150
    k_style.spacing = 0.0
    subs.styles["Karaoke"] = k_style

    # --- Tạo event cho từng dòng: CHỈ MỘT event karaoke mỗi dòng ---
    for words_info in subtitles_lines:
        built = _build_karaoke_line(words_info)
        if not built:
            continue
        line_start, line_end, full_text, words = built

        # Xây text với \k tags (centiseconds)
        kara_text_parts: List[str] = []
        for idx, w in enumerate(words):
            # duration của từ (end_rel - start_rel) -> centiseconds
            dur_cs_word = max(1, int(round((w["end_rel"] - w["start_rel"]) * 100)))
            # escape { and } để tránh phá tag ASS
            safe_word = w["word"].replace("{", "\\{").replace("}", "\\}")
            kara_text_parts.append(f"{{\\k{dur_cs_word}}}{safe_word}")

            # xử lý gap giữa từ hiện tại và từ tiếp theo (nếu có)
            if idx + 1 < len(words):
                next_w = words[idx + 1]
                gap = next_w["start_rel"] - w["end_rel"]
                if gap > 0.001:
                    dur_cs_gap = max(1, int(round(gap * 100)))
                    # thêm một ký tự space được điều khiển bằng \k để giữ timing
                    kara_text_parts.append(f"{{\\k{dur_cs_gap}}} ")

        kara_text = "".join(kara_text_parts)

        fg_start = pysubs2.make_time(s=line_start)
        fg_end = pysubs2.make_time(s=line_end)
        fg_event = pysubs2.SSAEvent(start=fg_start, end=fg_end, text=kara_text, style="Karaoke")
        subs.events.append(fg_event)

    # Lưu file ASS
    subs.save(subtitle_path)

    # Render bằng ffmpeg (escape path an toàn)
    # Sử dụng list args để tránh shell injection
    vf_arg = f"ass={subtitle_path}"
    ffmpeg_cmd = [
        "ffmpeg", "-y",
        "-i", clip_video_path,
        "-vf", vf_arg,
        "-c:v", "libx264", "-preset", "fast", "-crf", "23",
        "-c:a", "copy",
        output_path
    ]

    subprocess.run(ffmpeg_cmd, check=True)

@app.cls(gpu="L40S", timeout=900, retries=0, scaledown_window=20, secrets=[modal.Secret.from_name("ai-podcast-clipper-secret")], volumes={mount_path: volume})
class AiPodcastClipper:
    @modal.enter()
    def load_model(self):
        from supabase import create_client, Client

        print("Loading models")

        self.whisperx_model = whisperx.load_model(
            "large-v2", device="cuda", compute_type="float16")

        self.alignment_model, self.metadata = whisperx.load_align_model(
            language_code="en",
            device="cuda"
        )

        print("Transcription models loaded...")

        print("Creating gemini client...")
        self.gemini_client = genai.Client(api_key=os.environ["GEMINI_API_KEY"])
        print("Created gemini client...")


        print("Creating Supabase client...")
        supabase_url: str = os.environ.get("SUPABASE_URL")
        supabase_key: str = os.environ.get("SUPABASE_SERVICE_ROLE_KEY")
        self.supabase_bucket_name: str = os.environ.get("SUPABASE_BUCKET_NAME")

        if not supabase_url or not supabase_key or not self.supabase_bucket_name:
             raise ValueError("Supabase URL, Service Role Key, or Bucket Name not found in environment variables/secrets.")
    
        self.supabase_client: Client = create_client(supabase_url, supabase_key)
        print("Created Supabase client.")

    def process_clip(self, base_dir: str, original_video_path: str, start_time: float, end_time: float, clip_index: int, transcript_segments: list, storage_path: str, config: VideoConfig ):
        # Extract unique identifier from storage_path to prevent collisions
        # storage_path format: "user-xxx/uuid-filename.mp4"
        path_parts = storage_path.split("/")
        filename_with_ext = path_parts[-1]
        file_identifier = os.path.splitext(filename_with_ext)[0] # "uuid-filename"

        clip_name = f"{file_identifier}_clip_{clip_index}"
        
        # Extract folder from original storage_path (e.g., "user-xxx/uuid-video.mp4" -> "user-xxx")
        storage_folder = "/".join(path_parts[:-1])
        output_s3_key = f"{storage_folder}/clips/{clip_name}.mp4"
        
      

        clip_dir = base_dir / clip_name
        clip_dir.mkdir(parents=True, exist_ok=True)

        clip_segment_path = clip_dir / f"{clip_name}_segment.mp4"
        vertical_mp4_path = clip_dir / "pyavi" / "video_out_vertical.mp4"
        subtitle_output_path = clip_dir / "pyavi" / "video_with_subtitles.mp4"

        (clip_dir / "pywork").mkdir(exist_ok=True)
        pyframes_path = clip_dir / "pyframes"
        pyavi_path = clip_dir / "pyavi"
        audio_path = clip_dir / "pyavi" / "audio.wav"

        pyframes_path.mkdir(exist_ok=True)
        pyavi_path.mkdir(exist_ok=True)

        duration = end_time - start_time
        cut_command = (f"ffmpeg -i {original_video_path} -ss {start_time} -t {duration} "
                    f"{clip_segment_path}")
        subprocess.run(cut_command, shell=True, check=True,
                    capture_output=True, text=True)

        extract_cmd = f"ffmpeg -i {clip_segment_path} -vn -acodec pcm_s16le -ar 16000 -ac 1 {audio_path}"
        subprocess.run(extract_cmd, shell=True,
                    check=True, capture_output=True)

        shutil.copy(clip_segment_path, base_dir / f"{clip_name}.mp4")

        columbia_command = (f"python Columbia_test.py --videoName {clip_name} "
                            f"--videoFolder {str(base_dir)} "
                            f"--pretrainModel weight/finetuning_TalkSet.model")

        columbia_start_time = time.time()
        process = subprocess.run(columbia_command, cwd="/asd", shell=True, capture_output=True, text=True)
        columbia_end_time = time.time()
        print(
            f"Columbia script completed in {columbia_end_time - columbia_start_time:.2f} seconds")
        # print("=== Columbia command output ===")
        # print(process.stdout)

        # print("=== Columbia command errors ===")
        # print(process.stderr)

        # print("Exit code:", process.returncode)

        tracks_path = clip_dir / "pywork" / "tracks.pckl"
        scores_path = clip_dir / "pywork" / "scores.pckl"
        if not tracks_path.exists() or not scores_path.exists():
            raise FileNotFoundError("Tracks or scores not found for clip")

        with open(tracks_path, "rb") as f:
            tracks = pickle.load(f)

        with open(scores_path, "rb") as f:
            scores = pickle.load(f)

        cvv_start_time = time.time()
        create_vertical_video(
            tracks, scores, pyframes_path, pyavi_path, audio_path, vertical_mp4_path, target_width=config.target_width, target_height=config.target_height
        )
        cvv_end_time = time.time()
        print(
            f"Clip {clip_index} vertical video creation time: {cvv_end_time - cvv_start_time:.2f} seconds")
        

        if config.subtitle:
            create_subtitles_with_ffmpeg(transcript_segments, start_time,
                                    end_time, vertical_mp4_path, subtitle_output_path, max_words=5)
            upload_path = subtitle_output_path
        else:
            upload_path = vertical_mp4_path

        with open(upload_path, "rb") as f:
            res = self.supabase_client.storage.from_(self.supabase_bucket_name).upload(
                output_s3_key, f
            )

        # Kiểm tra kết quả
        # if res.get("error"):
        #     raise Exception(f"Upload failed: {res['error']}")
        # else:
        #     print("Uploaded to Supabase:", output_s3_key)


    def transcribe_video(self, base_dir: str, video_path: str) -> str:
        audio_path = base_dir / "audio.wav"
        extract_cmd = f"ffmpeg -i {video_path} -vn -acodec pcm_s16le -ar 16000 -ac 1 {audio_path}"
        subprocess.run(extract_cmd, shell=True,
                       check=True, capture_output=True)

        print("Starting transcription with WhisperX...")
        start_time = time.time()

        audio = whisperx.load_audio(str(audio_path))
        result = self.whisperx_model.transcribe(audio, batch_size=16)

        result = whisperx.align(
            result["segments"],
            self.alignment_model,
            self.metadata,
            audio,
            device="cuda",
            return_char_alignments=False
        )

        duration = time.time() - start_time
  
        print("Transcription and alignment took " + str(duration) + " seconds")

        segments = []

        if "word_segments" in result:
            for word_segment in result["word_segments"]:
                segments.append({
                    "start": word_segment["start"],
                    "end": word_segment["end"],
                    "word": word_segment["word"],
                })

        return json.dumps(segments)
    
                                
    def identify_moments(self, transcript: dict, prompt: str):

        prompt_content=f"""
        This is a podcast video transcript consisting of words, along with each word's start and end time. I am looking to create clips based on a specific topic provided by the user.

        The user is specifically interested in moments related to: "{prompt}"

        Your task is to find and extract segments from the transcript that are relevant to the user's topic: "{prompt}". These segments could be stories, discussions, questions and answers, or significant mentions related to the topic.

        Each extracted clip must adhere strictly to the following rules:
        - The content must be directly relevant to the user's topic: "{prompt}".
        - Clip duration must be between a minimum of 30 seconds and a maximum of 60 seconds. Clips must never exceed 60 seconds.
        - Ensure that clips do not overlap with one another.
        - Start and end timestamps of the clips must align perfectly with the word boundaries in the transcript provided. Only use the start and end timestamps provided in the input; modifying timestamps is not allowed.
        - Format the output STRICTLY as a list of JSON objects, each representing a clip with 'start' and 'end' timestamps in seconds: [{{"start": seconds, "end": seconds}}, ...clip2, clip3]. The output must be readable by the python json.loads function.
        - Aim to generate longer clips (closer to 40-60 seconds) that capture a complete thought or segment related to the user's topic, including relevant context if necessary.
        Avoid including:
        - Moments of greeting, thanking, or saying goodbye unless directly relevant to the user's topic.
        - Segments that are irrelevant to the user's topic: "{prompt}".

        If there are no valid clips relevant to "{prompt}" that meet all the criteria (especially duration), the output must be an empty list [], in JSON format, readable by json.loads() in Python.

        The transcript is as follows:\n\n""" + str(transcript)

                                  
       # --- Retry & Fallback Logic ---
        models_to_try = ["gemini-3-pro-preview","gemini-2.5-pro","gemini-2.5-flash", "gemini-2.5-flash-preview-09-2025","gemini-2.5-flash-lite"]
        max_retries_per_model = 3
        retry_delay_seconds = 5 
        
        last_error = None

        for model_name in models_to_try:
            print(f"Trying model: {model_name}")
            for attempt in range(max_retries_per_model):
                try:
                    # --- API Call ---
                    response = self.gemini_client.models.generate_content(
                        model=model_name, 
                        contents=prompt_content
                    )
                    
                    # --- Success ---
                    print(f"Successfully generated content using {model_name}")
                    return response.text 

                except google.genai.errors.ServerError as e:
                    # --- Lỗi Server (503) ---
                    print(f"Error with {model_name} (Attempt {attempt + 1}/{max_retries_per_model}): {e}")
                    last_error = e
                    if attempt < max_retries_per_model - 1:
                        print(f"Retrying in {retry_delay_seconds} seconds...")
                        time.sleep(retry_delay_seconds)
                
                except Exception as e:
                    # --- Lỗi khác (400, Auth...) ---
                    print(f"Unexpected error with {model_name}: {e}")
                    last_error = e
                    # Với lỗi không phải ServerError, break để thử model khác ngay
                    break 
            
            print(f"Failed with {model_name}. Switching to next model...")

        # Nếu chạy hết các model mà vẫn không return
        raise Exception(f"All models failed. Last error: {last_error}")


    @modal.fastapi_endpoint(method="POST")
    def process_video(self, request: ProcessVideoRequest, token: HTTPAuthorizationCredentials = Depends(auth_scheme)):
        
        api_secret = os.environ.get("API_SECRET")
        if api_secret is None:
             print("Warning: API_SECRET environment variable is not set. Skipping token validation.")
        elif token.credentials != api_secret:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Incorrect Bearer token",
                headers={"WWW-Authenticate": "Bearer"},
            )

        storage_path = request.storage_path
        video_config = request.config
        print(f"Received request to process video at storage path: {storage_path} with config: {video_config}")

        run_id = str(uuid.uuid4())
        base_dir = pathlib.Path(f"/tmp/{run_id}")
        base_dir.mkdir(parents=True, exist_ok=True)

        video_path = base_dir / "input_video.mp4"
        # Download video file from Supabase
        try:
            with open(video_path, 'wb+') as f:
                res = self.supabase_client.storage.from_(self.supabase_bucket_name).download(storage_path)
                f.write(res)
            print(f"Successfully downloaded {storage_path} to {video_path}")
        except Exception as e:
            print(f"Error downloading from Supabase: {e}")
            if base_dir.exists():
                shutil.rmtree(base_dir, ignore_errors=True)
            raise HTTPException(status_code=status.HTTP_500_INTERNAL_SERVER_ERROR, detail=f"Failed to download video from storage: {e}")
        
        # 1. Transcription
        transcript_segments_json = self.transcribe_video(base_dir, video_path)
        transcript_segments = json.loads(transcript_segments_json)
    
        
        # 2. Identify moments for clips
        print("Identifying clip moments")
        identified_moments_raw = self.identify_moments(transcript_segments, video_config.prompt)
        
        cleaned_json_string = identified_moments_raw.strip()
        if cleaned_json_string.startswith("```json"):
            cleaned_json_string = cleaned_json_string[len("```json"):].strip()
        if cleaned_json_string.endswith("```"):
            cleaned_json_string = cleaned_json_string[:-len("```")].strip()

        clip_moments = json.loads(cleaned_json_string)
        if not clip_moments or not isinstance(clip_moments, list):
            print("Error: Identified moments is not a list")
            clip_moments = []

        print(clip_moments)

         # 3. Process clips
        for index, moment in enumerate(clip_moments[:video_config.clip_count]):
            if "start" in moment and "end" in moment:
                print("Processing clip" + str(index) + " from " +
                      str(moment["start"]) + " to " + str(moment["end"]))
                self.process_clip( base_dir, video_path,
                             moment["start"], moment["end"], index, transcript_segments, storage_path, video_config)

        if base_dir.exists():
            print(f"Cleaning up temp dir after {base_dir}")
            shutil.rmtree(base_dir, ignore_errors=True)


            
@app.local_entrypoint()
def main():
    import requests

    ai_podcast_clipper = AiPodcastClipper()

    url = ai_podcast_clipper.process_video.get_web_url()                 

    payload = {
        "storage_path": "test/aa aaaaa aa.mp4",
        "config": {
            "prompt": "interesting topics about mi6",
            "clip_count": 1,
            "target_width": 1230,
            "target_height": 2230,
            "subtitle": False
        }
    }

    headers = {
        "Content-Type": "application/json",
        "Authorization": "Bearer 123123"
    }

    response = requests.post(url, json=payload,
                             headers=headers)
    response.raise_for_status()
    result = response.json()
    print(result)




 # prompt_content="""This is a podcast video transcript consisting of word, along with each words's start and end time. I am looking to create clips between a minimum of 30 and maximum of 60 seconds long. The clip should never exceed 60 seconds.

        # Your task is to find and extract stories, or question and their corresponding answers from the transcript.
        # Each clip should begin with the question and conclude with the answer.
        # It is acceptable for the clip to include a few additional sentences before a question if it aids in contextualizing the question.

        # Please adhere to the following rules:
        # - Ensure that clips do not overlap with one another.
        # - Start and end timestamps of the clips should align perfectly with the sentence boundaries in the transcript.
        # - Only use the start and end timestamps provided in the input. modifying timestamps is not allowed.
        # - Format the output as a list of JSON objects, each representing a clip with 'start' and 'end' timestamps: [{"start": seconds, "end": seconds}, ...clip2, clip3]. The output should always be readable by the python json.loads function.
        # - Aim to generate longer clips between 40-60 seconds, and ensure to include as much content from the context as viable.

        # Avoid including:
        # - Moments of greeting, thanking, or saying goodbye.
        # - Non-question and answer interactions.

        # If there are no valid clips to extract, the output should be an empty list [], in JSON format. Also readable by json.loads() in Python.
        # The transcript is as follows:\n\n""" + str(transcript)