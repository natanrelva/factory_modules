#!/usr/bin/env python3
"""
Windows TTS usando SAPI (Speech API)
Alternativa ao eSpeak para Windows
"""

import sys
import os

try:
    import pyttsx3
    HAS_PYTTSX3 = True
except ImportError:
    HAS_PYTTSX3 = False

def speak_sapi(text, output_file=None, rate=150, volume=1.0):
    """
    Sintetiza texto usando Windows SAPI
    
    Args:
        text: Texto para sintetizar
        output_file: Arquivo WAV de saída (opcional)
        rate: Velocidade (palavras por minuto)
        volume: Volume (0.0 a 1.0)
    """
    if not HAS_PYTTSX3:
        print("ERROR: pyttsx3 not installed. Install with: pip install pyttsx3", file=sys.stderr)
        return False
    
    try:
        engine = pyttsx3.init()
        
        # Configurar voz
        voices = engine.getProperty('voices')
        # Tentar usar voz em inglês
        for voice in voices:
            if 'english' in voice.name.lower() or 'en' in voice.id.lower():
                engine.setProperty('voice', voice.id)
                break
        
        # Configurar velocidade (150-200 wpm é bom)
        engine.setProperty('rate', rate)
        
        # Configurar volume (0.0 a 1.0)
        engine.setProperty('volume', volume)
        
        if output_file:
            # Salvar em arquivo
            engine.save_to_file(text, output_file)
            engine.runAndWait()
            print(f"[OK] Audio saved to: {output_file}")
        else:
            # Reproduzir diretamente
            engine.say(text)
            engine.runAndWait()
            print(f"[OK] Spoke: {text}")
        
        return True
        
    except Exception as e:
        print(f"ERROR: {e}", file=sys.stderr)
        return False

def main():
    if len(sys.argv) < 2:
        print("Usage: python windows-tts.py <text> [output.wav] [rate] [volume]")
        print("Example: python windows-tts.py \"Hello world\" output.wav 150 1.0")
        sys.exit(1)
    
    text = sys.argv[1]
    output_file = sys.argv[2] if len(sys.argv) > 2 else None
    rate = int(sys.argv[3]) if len(sys.argv) > 3 else 150
    volume = float(sys.argv[4]) if len(sys.argv) > 4 else 1.0
    
    success = speak_sapi(text, output_file, rate, volume)
    sys.exit(0 if success else 1)

if __name__ == "__main__":
    main()
