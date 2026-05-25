package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/i247app/gex/jwtutil"
	gexsess "github.com/i247app/gex/session"
	"github.com/i247app/gex/sessionprovider"

	"math-ai.com/math-ai/internal/adapter/bot"
	"math-ai.com/math-ai/internal/adapter/email"
	"math-ai.com/math-ai/internal/adapter/otp_delivery"
	"math-ai.com/math-ai/internal/adapter/sms"
	"math-ai.com/math-ai/internal/adapter/storage"
	"math-ai.com/math-ai/internal/application/resource"
	jobruntime "math-ai.com/math-ai/internal/infrastructure/job"
	"math-ai.com/math-ai/internal/infrastructure/logger"
	"math-ai.com/math-ai/internal/infrastructure/session"
)

const (
	sessionTTL = 14 * 24 * time.Hour // 14 days
)

func SetupResource(res *resource.Resource) error {
	log := logger.From(context.Background())

	env := res.Env

	log.Info("> Setup JobRegistry + JobRuntime...")
	res.JobRegistry = jobruntime.NewRegistry()
	res.JobRuntime = jobruntime.NewRuntime(jobruntime.Config{}, res.JobRegistry)

	log.Info("> Setup SessionManager...")
	sessionManager := session.NewSessionManager()
	res.SessionManager = sessionManager

	log.Info("> jwtHelper...")
	var jwtHelper jwtutil.JwtHelper
	if env.SharedKeyBytes != nil {
		helper, err := jwtutil.NewHmacJwtHelper(env.SharedKeyBytes)
		if err != nil {
			panic("failed to create jwt toolkit from env shared key")
		}
		jwtHelper = helper
	} else {
		return fmt.Errorf("unable to determine jwt helper from env")
	}

	var sessionProvider sessionprovider.SessionProvider
	{
		defaultSessFactory := func() gexsess.SessionStorer {
			// Create the basic session that all new sessions are based on
			return session.NewSession()
		}
		if env.GexSessionDriver == "xwt" {
			sessionProvider = sessionprovider.NewXwtSessionProvider(
				sessionManager.Container(),
				jwtHelper,
				defaultSessFactory,
				sessionTTL,
			)
		} else {
			sessionProvider = sessionprovider.NewJwtSessionProvider(
				sessionManager.Container(),
				jwtHelper,
				defaultSessFactory,
				sessionTTL,
			)
		}
	}
	res.SessionProvider = sessionProvider

	log.Info("> Setup Resource...")
	log.Info("> Setup EmailAdapter...")
	log.Infof("> Email Provider: %s", env.EmailConfig.EmailProvider)
	emailAdapter, err := email.NewFromConfig(context.Background(), env.EmailConfig)
	if err != nil {
		return fmt.Errorf("failed to setup email adapter: %w", err)
	}
	res.EmailProvider = emailAdapter

	log.Info("> Setup SMSAdapter...")
	log.Infof("> SMS Provider: %s", env.SMSConfig.SMSProvider)
	smsAdapter, err := sms.NewFromConfig(context.Background(), env.SMSConfig)
	if err != nil {
		return fmt.Errorf("failed to setup sms adapter: %w", err)
	}
	res.SMSProvider = smsAdapter

	log.Info("> Setup StorageAdapter...")
	log.Infof("> Storage Provider: %s", env.StorageConfig.Provider)
	storageAdapter, err := storage.NewFromConfig(context.Background(), env.StorageConfig)
	if err != nil {
		return fmt.Errorf("failed to setup storage adapter: %w", err)
	}
	res.StorageProvider = storageAdapter

	log.Info("> Setup BotAdapter...")
	log.Infof("> Bot Provider: %s (langchain backend: %s)",
		env.BotConfig.BotProvider, env.BotConfig.LangChainBackend)
	botAdapter, err := bot.NewFromConfig(context.Background(), env.BotConfig)
	if err != nil {
		return fmt.Errorf("failed to setup bot adapter: %w", err)
	}
	res.BotProvider = botAdapter

	log.Info("> Setup OtpDeliveryAdapter...")
	res.OtpDelivery = otp_delivery.NewFromAdapters(context.Background(), smsAdapter, emailAdapter)

	return nil
}
