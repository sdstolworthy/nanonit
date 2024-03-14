import time
import gc
import gifio
from os import getenv
import ssl
import socketpool
import wifi
from math import sin
import board
import displayio
import rgbmatrix
import framebufferio
import adafruit_imageload
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
    bitmap = displayio.Bitmap(64, 64, 1)
    for i in range(64):
        for j in range(64):
            bitmap[i, j] = 0
    return bitmap, palette

def initialize_matrix():
    displayio.release_displays()
    bit_depth_value = 4
    unit_width = 64
    unit_height = 64
    chain_width = 1
    chain_height = 1
    serpentine_value = True
    width_value = unit_width * chain_width
    height_value = unit_height * chain_height
    return rgbmatrix.RGBMatrix(
        width=width_value,
        height=height_value,
        bit_depth=bit_depth_value,
        rgb_pins=[board.GP2, board.GP3, board.GP4, board.GP5, board.GP8, board.GP9],
        addr_pins=[board.GP10, board.GP16, board.GP18, board.GP20, board.GP22],
        clock_pin=board.GP11,
        latch_pin=board.GP12,
        output_enable_pin=board.GP13,
        tile=chain_height,
        serpentine=serpentine_value,
        doublebuffer=True,
    )

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
    odg = gifio.OnDiskGif('/image.gif')
    t = displayio.TileGrid(odg.bitmap, pixel_shader=displayio.ColorConverter())
    group.append(t)
    return odg

main()
