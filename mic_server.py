#!/usr/bin/env python

import pyaudio
import socket
from threading import Thread

frames = []


def udpStream():
    udp = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    udp.setsockopt(socket.SOL_SOCKET, socket.SO_BROADCAST, 1)
    while True:
        if len(frames) > 0:
            udp.sendto(frames.pop(0), ("<broadcast>", 12345))

    udp.close()


def record(stream, CHUNK):
    while True:
        frames.append(stream.read(CHUNK))


CHUNK = 256
FORMAT = pyaudio.paInt16
CHANNELS = 1
RATE = 44100

p = pyaudio.PyAudio()

stream = p.open(format=FORMAT, channels=CHANNELS, rate=RATE, input=True, frames_per_buffer=CHUNK)


Tr = Thread(target=record, args=(stream, CHUNK))
Ts = Thread(target=udpStream)
Tr.setDaemon(True)
Ts.setDaemon(True)
Tr.start()
Ts.start()
Tr.join()
Ts.join()
