from io import BytesIO
from os import getenv
import time
import ssl
import wifi
import displayio
import microcontroller
import framebufferio
import rgbmatrix
import adafruit_imageload
import socketpool
import adafruit_requests
import board


IMAGE_REFRESH_FREQUENCY_IN_SECONDS = 5
TARGET_FPS = 30


def get_endpoint(device_id: str):
    base_url = getenv("base_url")
    return "%s/render/%s" % (base_url, device_id)


def get_image(device_id: str):
    try:
        print("getting image")
        print("device id %s" % device_id)
        radio = wifi.radio
        pool = socketpool.SocketPool(radio)
        ssl_context = ssl.create_default_context()
        requests = adafruit_requests.Session(pool, ssl_context)
        response = requests.get(get_endpoint(device_id))
        return BytesIO(response.content)
    except Exception as e:
        print("Error fetching image: ", e)
        return None


def get_device_id():
    cpuid = microcontroller.cpu.uid
    device_id = "".join("{:02x}".format(x) for x in cpuid)
    print(device_id)
    return device_id


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
    return matrix


#
#
def run_display(matrix):
    display = framebufferio.FramebufferDisplay(matrix, auto_refresh=False)
    
    g = displayio.Group()
    bitmap = get_image(get_device_id())
    b, p = adafruit_imageload.load(bitmap)
    print("loaded image")
    t = displayio.TileGrid(b, pixel_shader=p)
    g.append(t)


    display.root_group = g

    target_fps = 50
    ft = 1/target_fps
    now = t0 = time.monotonic_ns()
    deadline = t0 + ft

    while True:
        display.refresh(target_frames_per_second=target_fps, minimum_frames_per_second=0)
        while True:
            now = time.monotonic_ns()
            if now > deadline:
                break
            time.sleep((deadline - now) * 1e-9)
        deadline += ft



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


def main():
    matrix = initialize_matrix()
    initialize_wifi()
    run_display(matrix)


main()
