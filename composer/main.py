import logging
import json
import os
import dataclasses
import random
from typing import List
from PIL import Image, ImageDraw, ImageFont
import math
from type import MovieMeta, ScriptItem
import glob
from itertools import chain

from moviepy import (
    VideoClip,
    VideoFileClip,
    ImageSequenceClip,
    ImageClip,
    TextClip,
    ColorClip,
    AudioFileClip,
    AudioClip,
    UpdatedVideoClip,
    CompositeVideoClip,
    concatenate_videoclips,
    concatenate_audioclips,
    clips_array,
    vfx,
)
import numpy as np

log = logging.getLogger(__name__)

workdir = os.getenv("WORKDIR", "/default/workdir")
metafile = os.getenv("METAFILE", "meta.json")
env = os.getenv("ENV", "dev")

ROOT = os.path.dirname(os.path.abspath(__file__))

# prepare_data function to convert raw data into a MovieMeta object
def prepare_data(data: dict) -> MovieMeta:
    script_items = data.get("script_items", [])

    meta = MovieMeta(**data)
    meta.script_items = [ScriptItem(**item)
            for item in script_items]
    return meta

def validate_data(meta: MovieMeta) -> bool:
    for item in meta.script_items:
        if not item.cn or not item.en or not item.image_prompt or not item.voice_path or not item.image_path:
            log.error(f"Invalid script item: {item}")
            return False

    for item in meta.script_items:
        if not os.path.exists(item.voice_path):
            log.error(f"Voice file does not exist: {item.voice_path}")
            return False
        if not os.path.exists(item.image_path):
            log.error(f"Image file does not exist: {item.image_path}")
            return False

# Define some constants for later use
black = (0, 0, 0)  # RGB for black
red = (255, 0, 0)  # RGB for red
gray = (128, 128, 128)  # RGB for gray
light_gray = (211, 211, 211)  # RGB for light gray
white = (255, 255, 255)  # RGB for white

# resolution settings
height = 1920
width = 1080
hgihtlight_height = 1920 * (1 - 0.618)
padding = 20

origin = (0,  height / 2 - hgihtlight_height / 2) #left_top
center = (width / 2, height / 2)
bottom_left = (0, int(height / 2 + hgihtlight_height / 2)) #left_bottom
top_right = (width, height / 2 - hgihtlight_height / 2) #right_top
bottom_right = (width, height / 2 + hgihtlight_height / 2)

# main text font  size
font_size_h1 = 50
font_height_h1 = int(font_size_h1 * 1.2)  # Adjust height based on font size

font_size_h2 = 30
font_height_h2 = int(font_size_h1 * 1.2)  # Adjust height based on font size

# audio settings
audio_pause = 0.3

silence_clip = AudioClip(lambda t: 0, duration=audio_pause)


def generate_video(meta: MovieMeta) -> str:
    audio_clips = []
    total_duration = 0
    start_time = 0
    for i, item in enumerate(meta.script_items):
        print(f"{workdir}/{item.voice_path}")
        audio_clip = AudioFileClip(f"{workdir}/{item.voice_path}")
        print(f"Processing audio for item {i+1}: {item.voice_path} {audio_clip.duration} seconds")
        audio_clip = audio_clip.with_volume_scaled(0.9)
        audio_clip = audio_clip.with_start(start_time + i * audio_pause)
        audio_clip = audio_clip.with_end(start_time + i * audio_pause + audio_clip.duration)
        audio_clips.append(audio_clip)
        total_duration += audio_clip.duration
    audio_clip = concatenate_audioclips(list(chain.from_iterable([[ac, silence_clip] for ac in audio_clips])))
    total_duration = total_duration + (len(meta.script_items) -1 )* audio_pause

    image_clips = []
    start_time = 0
    for i, item in enumerate(meta.script_items):
        duration = audio_clips[i].duration
        print(f"Processing image for item {i+1}: {item.image_path} ")
        image_clip = ImageClip(f"{workdir}/image/{i}.png")
        image_clip = image_clip.with_duration(duration + (len(meta.script_items) - 1) * audio_pause)
        image_clip = image_clip.with_start(start_time + i * audio_pause)
        image_clip = image_clip.with_end(start_time + i * audio_pause + duration )
        image_clip = image_clip.resized(new_size=(height/3, width/3))
        image_clips.append(image_clip)
        start_time += duration

    max_image_height = max(clip.size[1] for clip in image_clips)
    max_image_width = max(clip.size[0] for clip in image_clips)

    text_clips = []
    start_time = 0
    for i, item in enumerate(meta.script_items):
        print(f"Processing item {i+1}: {item.cn} / {item.en}")
        duration = audio_clips[i].duration
        text_clip = TextClip(text=item.cn, font=f"{ROOT}/fonts/ZCOOLQingKeHuangYou-Regular.ttf",
                             font_size=font_size_h1,
                             color=black,
                             size=(cal_text_width(item.cn, font_size_h1), font_height_h1),
                             bg_color=white)
        text_clip = text_clip.with_duration(duration + (len(meta.script_items) - 1) * audio_pause)
        text_clip = text_clip.with_start(start_time + i * audio_pause)
        text_clip = text_clip.with_end(start_time + i * audio_pause + duration)
        text_clips.append(text_clip)
        start_time += duration

    en_text_clips = []
    start_time = 0
    for i, item in enumerate(meta.script_items):
        duration = audio_clips[i].duration
        print(f"Processing item {i+1}: {item.en}")
        text_clip = TextClip(text=item.en,
                             font=f"{ROOT}/fonts/ZCOOLQingKeHuangYou-Regular.ttf",
                             color=black,
                             font_size=font_size_h2,
                             size=(cal_text_width(item.cn * 4, font_size_h1), font_height_h2),
                             bg_color=white)
        text_clip = text_clip.with_duration(duration + (len(meta.script_items) - 1) * audio_pause)
        text_clip = text_clip.with_start(start_time + i * audio_pause)
        text_clip = text_clip.with_end(start_time + + i * audio_pause + duration)
        en_text_clips.append(text_clip)
        start_time += duration

    bg_clip = ColorClip(size=(width, height), color=black, duration=total_duration)
    hightlight_clip = ColorClip(size=(width, int(hgihtlight_height)), color=white, duration=total_duration)

    print(f"bg_clip duration: {bg_clip.duration} seconds size: {bg_clip.size}")

    hightlight_clip = hightlight_clip.with_position(origin)
    text_clips = [CompositeVideoClip(([clip.with_effects([vfx.CrossFadeIn(0.3)])])) for clip in text_clips]
    text_clips = [clip.with_position(("center", bottom_right[1] - font_height_h1 - padding - font_height_h1)) for clip in text_clips]

    en_text_clips = [CompositeVideoClip(([clip.with_effects([vfx.CrossFadeIn(0.3)])])) for clip in en_text_clips]
    en_text_clips = [clip.with_position(("center", bottom_right[1] - font_height_h1 - padding)) for clip in en_text_clips]

    image_effects = [
        [vfx.SlideIn(0.3, "left")],
        [vfx.SlideIn(0.3, "top")],
        [vfx.SlideIn(0.3, "right")],
        [vfx.CrossFadeIn(0.3)],
        # [vfx.CrossFadeOut(0.5)],
    ]
    image_clips = [CompositeVideoClip(([clip.with_effects(random.choice(image_effects))])) for clip in image_clips]
    image_clips = [clip.with_position(("center", center[1] - max_image_height / 2 )) for clip in image_clips]

    title_clip =TextClip(text=meta.title,
                         font=f"{ROOT}/fonts/ZCOOLQingKeHuangYou-Regular.ttf",
                         color=black,
                         size=(cal_text_width(meta.title, font_size_h2), font_height_h2),
                         bg_color=white)

    title_clip = title_clip.with_duration(total_duration)
    title_clip = title_clip.with_position((padding, origin[1] + padding))

    color_clip = ColorClip(size=(50, 50), color=red, duration=total_duration)
    color_clip1 = ColorClip(size=(50, 50), color=red, duration=total_duration)

    color_clip = color_clip.with_position((top_right[0] - 50, top_right[1]))
    color_clip1 = color_clip1.with_position((0, int(bottom_left[1] - 50)))

    print(f"color_clip duration: {bottom_left}")

    print(f"Total duration: {total_duration} seconds")
    for i, item in enumerate(meta.script_items):
        print(f"audio {i} start_time {audio_clips[i].start} end_time {audio_clips[i].end} duration {audio_clips[i].duration}")
        print(f"text {i} start_time {text_clips[i].start} end_time {text_clips[i].end} duration {text_clips[i].duration}")
        print(f"image {i} start_time {image_clips[i].start} end_time {image_clips[i].end} duration {image_clips[i].duration}")
        print(" ==== ")




    all_clips = [bg_clip,
                 hightlight_clip,
                 title_clip,
                 color_clip,
                 color_clip1,
                 ] + text_clips + image_clips + en_text_clips
    [print(f"clip {i} duration: {clip.duration} seconds size: {clip.size}") for i, clip in enumerate(all_clips)]
    final_clip = CompositeVideoClip(all_clips)
    print(f"final_clip duration: {final_clip.duration} seconds size: {final_clip.size}")
    final_clip = final_clip.with_audio(audio_clip)
    if env == "dev":
        final_clip.preview()
    final_clip.write_videofile(f"{workdir}/output.mp4", fps=24, codec='libx264', audio_codec='aac')
    #



# fuction to calculate the size of the text based on font size, this is chinese text, so we need to use a different approach
def cal_text_width(text: str, font_size: int) -> tuple:
    return int(font_size * len(text) * 0.8)  # Adjust width based on text length


def main():
    log.setLevel(logging.DEBUG)
    log.info("Starting the main function of the composer module.")
    log.info(f"Working directory is set to: {workdir}")
    print(f"Working directory is set to: {workdir}")
    for file in glob.glob(f"{workdir}/*"):
        print(file)

    meta_path = os.path.join(workdir, metafile)
    if not os.path.exists(meta_path):
        log.error(f"Meta file {metafile} does not exist in {workdir}.")
        return

    with open(os.path.join(workdir, metafile), 'r') as f:
        data = json.loads(f.read())

    meta = prepare_data(data)

    if validate_data(meta):
        log.info("Data validation passed.")


    generate_video(meta)


main()
