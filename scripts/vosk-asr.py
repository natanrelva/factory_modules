#!/usr/bin/env python3
"""
Vosk ASR (Automatic Speech Recognition)
Reconhecimento de fala em português usando Vosk
"""

import sys
import os
import json
import wave
import struct

try:
    from vosk import Model, KaldiRecognizer, SetLogLevel
    SetLogLevel(-1)  # Suprimir logs do Vosk
    HAS_VOSK = True
except ImportError:
    HAS_VOSK = False

def transcribe_audio(audio_file, model_path, sample_rate=16000):
    """
    Transcreve áudio usando Vosk
    
    Args:
        audio_file: Arquivo WAV de entrada
        model_path: Caminho para o modelo Vosk
        sample_rate: Taxa de amostragem (16000 Hz)
    
    Returns:
        Texto transcrito
    """
    if not HAS_VOSK:
        print("ERROR: vosk not installed. Install with: pip install vosk", file=sys.stderr)
        return None
    
    if not os.path.exists(model_path):
        print(f"ERROR: Model not found at {model_path}", file=sys.stderr)
        return None
    
    try:
        # Carregar modelo
        model = Model(model_path)
        recognizer = KaldiRecognizer(model, sample_rate)
        recognizer.SetWords(True)
        
        # Ler arquivo WAV
        with wave.open(audio_file, "rb") as wf:
            # Verificar formato
            if wf.getnchannels() != 1:
                print("ERROR: Audio must be mono", file=sys.stderr)
                return None
            if wf.getsampwidth() != 2:
                print("ERROR: Audio must be 16-bit", file=sys.stderr)
                return None
            if wf.getframerate() != sample_rate:
                print(f"WARNING: Audio sample rate is {wf.getframerate()}, expected {sample_rate}", file=sys.stderr)
            
            # Processar áudio
            while True:
                data = wf.readframes(4000)
                if len(data) == 0:
                    break
                recognizer.AcceptWaveform(data)
            
            # Obter resultado final
            result = json.loads(recognizer.FinalResult())
            text = result.get("text", "")
            
            if text:
                print(f"[OK] Transcribed: {text}")
            else:
                print("[OK] No speech detected")
            
            return text
            
    except Exception as e:
        print(f"ERROR: {e}", file=sys.stderr)
        return None

def transcribe_samples(samples, model_path, sample_rate=16000):
    """
    Transcreve samples de áudio (float32) usando Vosk
    
    Args:
        samples: Lista de samples float32 (-1.0 a 1.0)
        model_path: Caminho para o modelo Vosk
        sample_rate: Taxa de amostragem (16000 Hz)
    
    Returns:
        Texto transcrito
    """
    if not HAS_VOSK:
        print("ERROR: vosk not installed. Install with: pip install vosk", file=sys.stderr)
        return None
    
    if not os.path.exists(model_path):
        print(f"ERROR: Model not found at {model_path}", file=sys.stderr)
        return None
    
    try:
        # Carregar modelo
        model = Model(model_path)
        recognizer = KaldiRecognizer(model, sample_rate)
        recognizer.SetWords(True)
        
        # Converter float32 para int16
        audio_data = bytearray()
        for sample in samples:
            # Clampar entre -1.0 e 1.0
            sample = max(-1.0, min(1.0, sample))
            # Converter para int16
            int_sample = int(sample * 32767)
            # Adicionar como bytes (little-endian)
            audio_data.extend(struct.pack('<h', int_sample))
        
        # Processar áudio em chunks
        chunk_size = 8000  # 4000 frames * 2 bytes
        for i in range(0, len(audio_data), chunk_size):
            chunk = bytes(audio_data[i:i+chunk_size])
            recognizer.AcceptWaveform(chunk)
        
        # Obter resultado final
        result = json.loads(recognizer.FinalResult())
        text = result.get("text", "")
        
        if text:
            print(f"[OK] Transcribed: {text}")
        else:
            print("[OK] No speech detected")
        
        return text
        
    except Exception as e:
        print(f"ERROR: {e}", file=sys.stderr)
        return None

def main():
    if len(sys.argv) < 3:
        print("Usage: python vosk-asr.py <audio_file.wav> <model_path> [sample_rate]")
        print("Example: python vosk-asr.py audio.wav models/vosk-model-small-pt-0.3 16000")
        sys.exit(1)
    
    audio_file = sys.argv[1]
    model_path = sys.argv[2]
    sample_rate = int(sys.argv[3]) if len(sys.argv) > 3 else 16000
    
    text = transcribe_audio(audio_file, model_path, sample_rate)
    
    if text:
        print(text)
        sys.exit(0)
    else:
        sys.exit(1)

if __name__ == "__main__":
    main()
