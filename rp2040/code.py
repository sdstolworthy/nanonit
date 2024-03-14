import time
import io
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


def get_image(device_id: str):
    print("getting image")
    print("device id %s" % device_id)
    radio = wifi.radio
    pool = socketpool.SocketPool(radio)
    ssl_context = ssl.create_default_context()
    requests = adafruit_requests.Session(pool, ssl_context)
    response = requests.get("http://10.0.0.37:8080/render/%s" % device_id)
    return io.BytesIO(response.content)

def get_device_id():
    cpuid = microcontroller.cpu.uid
    device_id = "".join("{:02x}".format(x) for x in cpuid)
    print(device_id)
    return device_id


def main():
    device_id = get_device_id()
    initialize_wifi()
    displayio.release_displays()
    bit_depth_value = 4
    unit_width = 64
    unit_height = 64
    chain_width = 1
    chain_height = 1
    serpentine_value = True
    width_value = unit_width * chain_width
    height_value = unit_height * chain_height
    matrix = rgbmatrix.RGBMatrix(
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
    display = framebufferio.FramebufferDisplay(matrix, auto_refresh=False)

    target_fps = 30
    refresh_frequency_in_minutes = 1 / target_fps
    now = t0 = time.monotonic_ns()
    screen_refresh_deadline = t0 + refresh_frequency_in_minutes
    image_refresh_frequency = 5 * 1e9
    image_refresh_deadline = t0 + image_refresh_frequency

    p = 1

    initialize_wifi()

    g = displayio.Group()
    while True:
        display.refresh(
            target_frames_per_second=target_fps, minimum_frames_per_second=0
        )
        now = time.monotonic_ns()
        if now > image_refresh_deadline:
            try:
                image = get_image(device_id)
                b, p = adafruit_imageload.load(image)
                t = displayio.TileGrid(b, pixel_shader=p)
                g.append(t)
                display.root_group = g
                image_refresh_deadline += image_refresh_frequency
            except:
                pass
        while True:
            now = time.monotonic_ns()
            if now > screen_refresh_deadline:
                break
            time.sleep((screen_refresh_deadline - now) * 1e-9)
        screen_refresh_deadline += refresh_frequency_in_minutes


main()
