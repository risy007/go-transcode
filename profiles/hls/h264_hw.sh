#!/bin/sh

#export VW="1920"
#export VH="1080"
export ABANDWIDTH="192k"
export VBANDWIDTH="5000k"
export VMAXRATE="5350k"
export VBUFSIZE="7500k"


exec ffmpeg \
  -y \
  -i "${1}" \
  -map 0:v:0 -map 0:a:0 \
  -c:a aac \
    -ar 48000 \
    -ac 2 \
    -b:a $ABANDWIDTH \
  -c:v h264_nvenc \
    -profile:v main \
    -force_key_frames "expr:gte(t,n_forced*1)" \
    -b:v $VBANDWIDTH \
    -maxrate $VMAXRATE \
    -bufsize $VBUFSIZE \
    -crf 20 \
    -sc_threshold 0 \
    -g 48 \
    -keyint_min 48 \
  -f hls \
    -hls_time 2 \
    -hls_list_size 5 \
    -hls_delete_threshold 1 \
    -hls_flags delete_segments+second_level_segment_index \
    -hls_start_number_source datetime \
    -strftime 1 \
    -hls_segment_filename "${2}/live_%03d.ts" \
    "${2}/index.m3u8" -
