# we have to use amd64 platform/arch because not all package are available for arm
FROM --platform=linux/amd64 jkassis/xgo:1.19.5
RUN apt-get update
RUN apt-get install -y --no-install-recommends --fix-missing libsdl2-2.0-0 libsdl2-dev libsdl2-image-dev libsdl2-mixer-dev libsdl2-ttf-dev pulseaudio
# RUN pkg-config --cflags --libs sdl2
# RUN sdl2-config --cflags
