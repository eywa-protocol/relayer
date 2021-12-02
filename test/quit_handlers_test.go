package test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.digiu.ai/blockchainlaboratory/eywa-p2p-bridge/node/bridge"
)

func Test_AddRemove(t *testing.T) {
	handlers := bridge.NewQuitHandlers()
	ch1 := handlers.AddQuitHandler("one")
	ch2 := handlers.AddQuitHandler("two")

	hl, tl := handlers.DebugMapLength()
	require.Equal(t, hl, 2)
	require.Equal(t, tl, 2)

	handlers.RemoveTopicHandler("one")

	select {
	case <-ch1:
	default:
		t.Error("First channel was NOT closed")
	}

	select {
	case <-ch2:
		t.Error("Second channel was closed")
	default:
	}

	handlers.RemoveTopicHandler("two")

	select {
	case <-ch2:
	default:
		t.Error("Second channel was NOT closed")
	}

	hl, tl = handlers.DebugMapLength()
	require.Zero(t, hl)
	require.Zero(t, tl)
}

func Test_Replace(t *testing.T) {
	handlers := bridge.NewQuitHandlers()
	ch := handlers.AddQuitHandler("one")

	// Replace previous, so ch must close
	ch2 := handlers.AddQuitHandler("one")

	select {
	case <-ch:
	default:
		t.Error("First channel was NOT closed")
	}

	select {
	case <-ch2:
		t.Error("Second channel was closed")
	default:
	}

	// Remove with old channel, so nothing must change
	res := handlers.RemoveQuitHandler(ch)
	require.Nil(t, res)

	select {
	case <-ch2:
		t.Error("Second channel was closed")
	default:
	}

	// Remove with the right channel, it must close
	res2 := handlers.RemoveQuitHandler(ch2)
	require.Equal(t, "one", *res2)

	select {
	case <-ch2:
	default:
		t.Error("Second channel was NOT closed")
	}

	hl, tl := handlers.DebugMapLength()
	require.Zero(t, hl)
	require.Zero(t, tl)
}
