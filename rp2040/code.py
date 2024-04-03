import time
import gc
import gifio
import adafruit_imageload
from os import getenv
import ssl
import socketpool
import wifi
from math import sin
import board
import displayio
import rgbmatrix
import framebufferio
import adafruit_requests
from adafruit_display_text.label import Label
import microcontroller

_session = None
_locked = False

def get_endpoint(device_id: str):
    base_url = getenv("base_url")
    return "%s/render/%s" % (base_url, device_id)


def initialize_wifi():
    secrets = {
        "ssid": getenv("CIRCUITPY_WIFI_SSID"),
        "password": getenv("CIRCUITPY_WIFI_PASSWORD"),
    }
    print("Connecting to AP...")
    while not wifi.radio.ipv4_address:
        try:
            wifi.radio.connect(secrets["ssid"], secrets["password"])
        except ConnectionError as e:
            print("could not connect to AP, retrying: ", e)
        except Exception as e:
            print("Something went wrong, retrying: ", e)
    print("Connected to", secrets["ssid"], "\nIP address:", wifi.radio.ipv4_address)


def get_session():
    global _session
    if _session is None:
        radio = wifi.radio
        pool = socketpool.SocketPool(radio)
        ssl_context = ssl.create_default_context()
        ssl_context.check_hostname = False
        _session = adafruit_requests.Session(pool, ssl_context)
    return _session

def should_get_new_image(device_id: str):
    global _locked
    if _locked:
        return
    _locked = True
    try:
        print("checking for new image")
        requests = get_session()
        with requests.head(get_endpoint(device_id)) as response:
            if response.status_code == 304:
                print("no new image")
                return False
    except Exception as e:
        print("error checking for new image: ", e)
        return False
    finally:
        _locked = False
        gc.collect()
    return True

def get_image(device_id: str):
    global _locked
    if _locked:
        return
    _locked = True
    try:
        print("getting image for device id: %s" % device_id)
        requests = get_session()
        with requests.get(get_endpoint(device_id)) as response:
            with open("image.gif", "wb") as f:
                for chunk in response.iter_content(1024):
                    f.write(chunk)
    except Exception as e:
        print("error getting image: ", e)
    finally:
        _locked = False
        gc.collect()


def get_device_id():
    cpuid = microcontroller.cpu.uid
    device_id = "".join("{:02x}".format(x) for x in cpuid)
    print("device id: %s" % device_id)
    return device_id

def blanking_bitmap():
    palette = displayio.Palette(1)
    palette[0] = 0x000000
    bitmap = displayio.Bitmap(64, 32, 1)
    for i in range(64):
        for j in range(32):
            bitmap[i, j] = 0
    return bitmap, palette

def initialize_matrix():
    displayio.release_displays()
    return rgbmatrix.RGBMatrix(

      width=64, bit_depth=3,

      rgb_pins=[board.GP0, board.GP1, board.GP2, board.GP3, board.GP5, board.GP4],

      addr_pins=[board.GP6, board.GP7, board.GP8, board.GP9],

      clock_pin=board.GP10, latch_pin=board.GP12, output_enable_pin=board.GP13)

def initialize_display(matrix):
    display = framebufferio.FramebufferDisplay(matrix, auto_refresh=True)
    g = displayio.Group()
    display.root_group = g
    return display, g

def main():
    device_id = get_device_id()
    initialize_wifi()
    matrix = initialize_matrix()
    _, g = initialize_display(matrix)

    gif = None
    try:
        gif = attach_image_to_group(g)
    except Exception as e:
        print(e)


    now = time.monotonic()
    next_image_check = now
    IMAGE_REFRESH_INTERVAL = 12
    next_frame_time = now

    while True:
        now = time.monotonic()
        if now >= next_image_check:
            try:
                if should_get_new_image(device_id):
                    if len(g) > 0:
                        g.pop()
                    get_image(device_id)
                    gif = attach_image_to_group(g)
            except Exception as e:
                print(e)
            next_image_check = time.monotonic() + IMAGE_REFRESH_INTERVAL
        if now > next_frame_time and gif is not None:
            delay = gif.next_frame()
            if delay is None:
                delay = 0.1
            next_frame_time = now + delay
        time.sleep(0.1)

def attach_image_to_group(group):
    # bitmap, palette = adafruit_imageload.load("penguino.bmp",
    #                                       bitmap=displayio.Bitmap,
    #                                       palette=displayio.Palette)
    # t = displayio.TileGrid(bitmap, pixel_shader = palette)
    odg = gifio.OnDiskGif('image.gif')
    t = displayio.TileGrid(odg.bitmap, pixel_shader=displayio.ColorConverter(input_colorspace=displayio.Colorspace.RGB565_SWAPPED))
    group.append(t)
    return odg

main()
# SPDX-FileCopyrightText: 2019 Carter Nelson for Adafruit Industries
#
# SPDX-License-Identifier: MIT


# displayio.release_displays()
# 
# matrix = rgbmatrix.RGBMatrix(
# 
#   width=64, bit_depth=2,
# 
#   rgb_pins=[board.GP0, board.GP1, board.GP2, board.GP3, board.GP5, board.GP4],
# 
#   addr_pins=[board.GP6, board.GP7, board.GP8, board.GP9],
# 
#   clock_pin=board.GP10, latch_pin=board.GP12, output_enable_pin=board.GP13)
# 
# display = framebufferio.FramebufferDisplay(matrix)
# bitmap, palette = adafruit_imageload.load("space.bmp",
#                                           bitmap=displayio.Bitmap,
#                                           palette=displayio.Palette)
# 
# # Create a TileGrid to hold the bitmap
# tile_grid = displayio.TileGrid(bitmap, pixel_shader=palette)
# 
# # Create a Group to hold the TileGrid
# group = displayio.Group()
# 
# # Add the TileGrid to the Group
# group.append(tile_grid)
# 
# # Add the Group to the Display
# display.root_group = group
# 
# # Loop forever so you can enjoy your image
# while True:
#     pass
# 
