#!/usr/bin/env python3
"""
Audio Capture usando PyAudio
Captura áudio real do microfone
"""

import sys
import os
import struct
import wave

try:
    import pyaudio
    HAS_PYAUDIO = True
except ImportError:
    HAS_PYAUDIO = False

def list_devices():
    """Lista dispositivos de áudio disponíveis"""
    if not HAS_PYAUDIO:
        print("ERROR: pyaudio not installed. Install with: pip install pyaudio", file=sys.stderr)
        return
    
    p = pyaudio.PyAudio()
    
    print("Available audio devices:")
    print("=" * 60)
    
    for i in range(p.get_device_count()):
        info = p.get_device_info_by_index(i)
        print(f"Device {i}: {info['name']}")
        print(f"  Max Input Channels: {info['maxInputChannels']}")
        print(f"  Max Output Channels: {info['maxOutputChannels']}")
        print(f"  Default Sample Rate: {info['defaultSampleRate']}")
        print()
    
    p.terminate()

def capture_audio(output_file, duration=3.0, sample_rate=16000, channels=1, device_index=None):
    """
    Captura áudio do microfone
    
    Args:
        output_file: Arquivo WAV de saída
        duration: Duração em segundos
        sample_rate: Taxa de amostragem (16000 Hz)
        channels: Número de canais (1 = mono)
        device_index: Índice do dispositivo (None = padrão)
    
    Returns:
        True se sucesso, False caso contrário
    """
    if not HAS_PYAUDIO:
        print("ERROR: pyaudio not installed. Install with: pip install pyaudio", file=sys.stderr)
        return False
    
    try:
        p = pyaudio.PyAudio()
        
        # Configurar stream
        stream = p.open(
            format=pyaudio.paInt16,
            channels=channels,
            rate=sample_rate,
            input=True,
            input_device_index=device_index,
            frames_per_buffer=1024
        )
        
        print(f"[OK] Recording {duration}s from microphone...")
        
        # Capturar áudio
        frames = []
        num_chunks = int(sample_rate / 1024 * duration)
        
        for i in range(num_chunks):
            data = stream.read(1024, exception_on_overflow=False)
            frames.append(data)
        
        print("[OK] Recording complete")
        
        # Parar stream
        stream.stop_stream()
        stream.close()
        p.terminate()
        
        # Salvar como WAV
        with wave.open(output_file, 'wb') as wf:
            wf.setnchannels(channels)
            wf.setsampwidth(p.get_sample_size(pyaudio.paInt16))
            wf.setframerate(sample_rate)
            wf.writeframes(b''.join(frames))
        
        print(f"[OK] Audio saved to: {output_file}")
        return True
        
    except Exception as e:
        print(f"ERROR: {e}", file=sys.stderr)
        return False

def capture_samples(duration=3.0, sample_rate=16000, channels=1, device_index=None):
    """
    Captura áudio e retorna samples float32
    
    Args:
        duration: Duração em segundos
        sample_rate: Taxa de amostragem (16000 Hz)
        channels: Número de canais (1 = mono)
        device_index: Índice do dispositivo (None = padrão)
    
    Returns:
        Lista de samples float32 (-1.0 a 1.0)
    """
    if not HAS_PYAUDIO:
        print("ERROR: pyaudio not installed. Install with: pip install pyaudio", file=sys.stderr)
        return None
    
    try:
        p = pyaudio.PyAudio()
        
        # Configurar stream
        stream = p.open(
            format=pyaudio.paInt16,
            channels=channels,
            rate=sample_rate,
            input=True,
            input_device_index=device_index,
            frames_per_buffer=1024
        )
        
        # Capturar áudio
        frames = []
        num_chunks = int(sample_rate / 1024 * duration)
        
        for i in range(num_chunks):
            data = stream.read(1024, exception_on_overflow=False)
            frames.append(data)
        
        # Parar stream
        stream.stop_stream()
        stream.close()
        p.terminate()
        
        # Converter para float32
        audio_data = b''.join(frames)
        samples = []
        
        for i in range(0, len(audio_data), 2):
            if i + 1 < len(audio_data):
                # Ler int16 (little-endian)
                int_sample = struct.unpack('<h', audio_data[i:i+2])[0]
                # Converter para float32 (-1.0 a 1.0)
                float_sample = int_sample / 32768.0
                samples.append(float_sample)
        
        print(f"[OK] Captured {len(samples)} samples")
        return samples
        
    except Exception as e:
        print(f"ERROR: {e}", file=sys.stderr)
        return None

def main():
    if len(sys.argv) < 2:
        print("Usage:")
        print("  python audio-capture.py list")
        print("  python audio-capture.py capture <output.wav> [duration] [sample_rate] [device_index]")
        print()
        print("Examples:")
        print("  python audio-capture.py list")
        print("  python audio-capture.py capture audio.wav 3 16000")
        print("  python audio-capture.py capture audio.wav 5 16000 1")
        sys.exit(1)
    
    command = sys.argv[1]
    
    if command == "list":
        list_devices()
        sys.exit(0)
    
    elif command == "capture":
        if len(sys.argv) < 3:
            print("ERROR: output file required", file=sys.stderr)
            sys.exit(1)
        
        output_file = sys.argv[2]
        duration = float(sys.argv[3]) if len(sys.argv) > 3 else 3.0
        sample_rate = int(sys.argv[4]) if len(sys.argv) > 4 else 16000
        device_index = int(sys.argv[5]) if len(sys.argv) > 5 else None
        
        success = capture_audio(output_file, duration, sample_rate, 1, device_index)
        sys.exit(0 if success else 1)
    
    else:
        print(f"ERROR: Unknown command: {command}", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()
