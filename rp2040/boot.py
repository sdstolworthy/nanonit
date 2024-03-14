# SPDX-FileCopyrightText: 2017 Limor Fried for Adafruit Industries
#
# SPDX-License-Identifier: MIT

"""CircuitPython Essentials Storage logging boot.py file"""
import board
import digitalio
import storage

switch = digitalio.DigitalInOut(board.GP0)

switch.direction = digitalio.Direction.INPUT
switch.pull = digitalio.Pull.UP

# Grounding GP0 will allow CircuitPython to write to the filesystem
storage.remount("/", readonly=switch.value)

