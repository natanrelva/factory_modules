package main

import (
	"fmt"
	"github.com/dubbing-system/audio-interface/pkg/coordinator"
	"github.com/dubbing-system/audio-interface/pkg/types"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	fmt.Println("🎙️  Audio Interface - Loopback Example")
	fmt.Println("=====================================")
	fmt.Println()

	// Configuração para voz (16kHz, mono, 20ms frames)
	config := types.AudioConfig{
		DeviceID:   "",     // Dispositivo padrão
		SampleRate: 16000,  // 16kHz para voz
		Channels:   1,      // Mono
		FrameSize:  320,    // 20ms @ 16kHz = 320 samples
		BufferSize: 10,     // 10 frames = 200ms de buffer
	}

	fmt.Println("📋 Configuração:")
	fmt.Printf("  Sample Rate: %d Hz\n", config.SampleRate)
	fmt.Printf("  Channels:    %d (Mono)\n", config.Channels)
	fmt.Printf("  Frame Size:  %d samples (%.1fms)\n", config.FrameSize, float64(config.FrameSize)/float64(config.SampleRate)*1000)
	fmt.Printf("  Buffer Size: %d frames (%dms)\n", config.BufferSize, config.BufferSize*20)
	fmt.Println()

	// Criar coordenador
	coord := coordinator.NewAudioInterfaceCoordinator(config)

	// Inicializar
	fmt.Print("🔧 Inicializando... ")
	if err := coord.Initialize(); err != nil {
		fmt.Printf("❌ Erro: %v\n", err)
		fmt.Println("\n💡 Dicas:")
		fmt.Println("  - Verifique se o microfone está conectado")
		fmt.Println("  - Verifique permissões de acesso ao microfone")
		fmt.Println("  - Feche outras aplicações usando áudio")
		return
	}
	fmt.Println("✅")

	// Iniciar captura e playback
	fmt.Print("▶️  Iniciando... ")
	if err := coord.Start(); err != nil {
		fmt.Printf("❌ Erro: %v\n", err)
		return
	}
	fmt.Println("✅")
	defer coord.Close()

	fmt.Println()
	fmt.Println("✅ Loopback ativo!")
	fmt.Println("🎤 Fale no microfone e ouça o eco nos alto-falantes")
	fmt.Println()
	fmt.Println("📊 Monitorando métricas...")
	fmt.Println("Pressione Ctrl+C para parar")
	fmt.Println()

	// Goroutine para exibir métricas
	go displayMetrics(coord)

	// Aguardar sinal de interrupção
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n\n🛑 Encerrando...")
	
	// Exibir sumário final
	displayFinalSummary(coord)
}

func displayMetrics(coord *coordinator.AudioInterfaceCoordinator) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Obter métricas
		latencyStats := coord.GetLatencyStats()
		syncStats := coord.GetSyncStats()
		summary := coord.GetMetricsSummary()

		// Limpar linha anterior
		fmt.Print("\r\033[K")

		// Exibir métricas em uma linha
		fmt.Printf("⏱️  Latência: %3dms | 📊 P95: %3dms | 🔄 Drift: %+3dms | ⚠️  Erros: %d | ⏰ Uptime: %v",
			latencyStats.EndToEndLatency.Milliseconds(),
			latencyStats.P95Latency.Milliseconds(),
			syncStats.DriftCompensation.Milliseconds(),
			summary.TotalErrors,
			summary.Uptime.Round(time.Second),
		)
	}
}

func displayFinalSummary(coord *coordinator.AudioInterfaceCoordinator) {
	summary := coord.GetMetricsSummary()
	latencyStats := coord.GetLatencyStats()

	fmt.Println("\n📈 Sumário Final:")
	fmt.Println("─────────────────────────────────────")
	fmt.Printf("  Uptime:           %v\n", summary.Uptime.Round(time.Second))
	fmt.Printf("  Módulos ativos:   %d\n", summary.TotalModules)
	fmt.Printf("  Total de erros:   %d\n", summary.TotalErrors)
	fmt.Println()
	fmt.Printf("  Latência média:   %v\n", latencyStats.AverageLatency)
	fmt.Printf("  Latência P50:     %v\n", latencyStats.P50Latency)
	fmt.Printf("  Latência P95:     %v\n", latencyStats.P95Latency)
	fmt.Printf("  Latência P99:     %v\n", latencyStats.P99Latency)
	fmt.Println()

	// Status
	avgMs := latencyStats.AverageLatency.Milliseconds()
	if avgMs <= 80 {
		fmt.Println("  Status: ✅ EXCELENTE (dentro do alvo)")
	} else if avgMs <= 100 {
		fmt.Println("  Status: ⚠️  BOM (próximo do alvo)")
	} else {
		fmt.Println("  Status: ❌ PRECISA OTIMIZAÇÃO")
	}

	fmt.Println("─────────────────────────────────────")
	fmt.Println("\n👋 Até logo!")
}
