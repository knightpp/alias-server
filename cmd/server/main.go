package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	accountsvc "github.com/knightpp/alias-proto/go/account/service/v1"
	gamesvc "github.com/knightpp/alias-proto/go/game/service/v1"
	accountserver "github.com/knightpp/alias-server/internal/accountserver"
	server "github.com/knightpp/alias-server/internal/gameserver"
	"github.com/knightpp/alias-server/internal/storage"
	"github.com/knightpp/alias-server/internal/storage/memory"
	"github.com/knightpp/alias-server/internal/storage/redis"
	"github.com/rs/zerolog"
	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
)

var (
	ngrokFlag     = flag.Bool("ngrok", false, "starts ngrok tunnel")
	ngrokAuthFlag = flag.String("ngrok-auth", "2Omz9oTCclkfVSwCFf8GBFsDt5E_7rmnvXs7aUePuNh8pGzmc", "auth token for ngrok")
	addr          string
	useH2C        bool
)

func init() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	h2c := false
	if os.Getenv("USE_H2C") == "1" {
		h2c = true
	}

	flag.StringVar(&addr, "addr", "0.0.0.0:"+port, "addr to listen to")
	flag.BoolVar(&useH2C, "h2c", h2c, "enables TLS")
}

func main() {
	flag.Parse()

	log := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().
		Timestamp().
		Logger().
		Level(zerolog.TraceLevel)

	if err := run(log); err != nil {
		log.Fatal().Err(err).Msg("server failed")
	}
}

func run(log zerolog.Logger) error {
	var playerDB storage.Player
	if url, ok := os.LookupEnv("REDIS_URL"); ok {
		pdb, err := redis.NewFromURL(url)
		if err != nil {
			return err
		}

		playerDB = pdb
	} else if addr, ok := os.LookupEnv("REDIS_ADDR"); ok {
		playerDB = redis.New(addr)
	} else {
		log.Warn().Msg("using inmem storage")
		playerDB = memory.New()
	}

	gameServer := server.New(log, playerDB)

	grpcLog := interceptorLogger(log)
	grpcServer := grpc.NewServer(
		grpc.ChainStreamInterceptor(
			logging.StreamServerInterceptor(grpcLog),
			recovery.StreamServerInterceptor(),
		),
		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(grpcLog),
			recovery.UnaryServerInterceptor(),
		),
	)
	gamesvc.RegisterGameServiceServer(grpcServer, gameServer)
	accountsvc.RegisterAccountServiceServer(grpcServer, accountserver.New(playerDB))

	log.Info().Str("addr", addr).Msg("starting GRPC server")

	if !useH2C {
		lis, err := listen(log, addr)
		if err != nil {
			return err
		}

		return grpcServer.Serve(lis)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(grpcServer.ServeHTTP))
	return http.ListenAndServe(
		addr,
		h2c.NewHandler(mux, &http2.Server{}),
	)
}

func interceptorLogger(l zerolog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l = l.With().Fields(fields).Logger()

		switch lvl {
		case logging.LevelDebug:
			l.Debug().Msg(msg)
		case logging.LevelInfo:
			l.Info().Msg(msg)
		case logging.LevelWarn:
			l.Warn().Msg(msg)
		case logging.LevelError:
			l.Error().Msg(msg)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}

func listen(log zerolog.Logger, addr string) (net.Listener, error) {
	if *ngrokFlag {
		log.Info().Msg("using ngrok")
		tun, err := ngrok.Listen(
			context.Background(),
			config.TCPEndpoint(),
			ngrok.WithAuthtoken(*ngrokAuthFlag),
		)
		if err != nil {
			return nil, fmt.Errorf("create ngrok tunnel: %w", err)
		}

		log.Info().Str("url", tun.URL()).Msg("started ngrok tunnel")
		return tun, nil
	} else {
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			return nil, fmt.Errorf("listen socket: %w", err)
		}
		return lis, nil
	}
}
