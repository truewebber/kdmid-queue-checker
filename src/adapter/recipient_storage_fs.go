package adapter

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"sync"

	"kdmid-queue-checker/domain/log"
	"kdmid-queue-checker/domain/notification"
)

type recipientStorageFs struct {
	storageFile  string
	cache        []notification.Recipient
	m            sync.RWMutex
	storageLimit int
	logger       log.Logger
}

type recipient struct {
	TelegramID int64  `json:"telegram_id"`
	ID         string `json:"id"`
	CD         string `json:"cd"`
}

func NewRecipientStorageFs(dir string, storageLimit uint8, logger log.Logger) (notification.Storage, error) {
	const storageFileName = "recipients.json"

	fs := &recipientStorageFs{
		storageFile:  path.Join(dir, storageFileName),
		storageLimit: int(storageLimit),
		logger:       logger,
	}

	err := fs.readAllToCache()

	if err != nil && !errors.Is(err, errNoFileExists) {
		return nil, fmt.Errorf("failed to read recipients from disk: %w", err)
	}

	if errors.Is(err, errNoFileExists) {
		if err := fs.createStorage(); err != nil {
			return nil, fmt.Errorf("failed to create storage: %w", err)
		}
	}

	return fs, nil
}

func MustNewRecipientStorageFs(dir string, storageLimit uint8, logger log.Logger) notification.Storage {
	storage, err := NewRecipientStorageFs(dir, storageLimit, logger)
	if err != nil {
		panic(err)
	}

	return storage
}

func (r *recipientStorageFs) Register(domainRecipient notification.Recipient) error {
	r.m.Lock()
	defer r.m.Unlock()

	if err := r.addRecipientIfNotPresented(domainRecipient); err != nil {
		return fmt.Errorf("failed to add recipient: %w", err)
	}

	if err := r.writeCache(); err != nil {
		return fmt.Errorf("failed to write recipients to disk: %w", err)
	}

	return nil
}

func (r *recipientStorageFs) Unregister(domainRecipient notification.Recipient) error {
	r.m.Lock()
	defer r.m.Unlock()

	if err := r.deleteRecipientIfPresented(domainRecipient); err != nil {
		return fmt.Errorf("failed to remove recipient: %w", err)
	}

	if err := r.writeCache(); err != nil {
		return fmt.Errorf("failed to write recipients to disk: %w", err)
	}

	return nil
}

func (r *recipientStorageFs) List() ([]notification.Recipient, error) {
	r.m.RLock()
	defer r.m.RUnlock()

	return r.cache, nil
}

func (r *recipientStorageFs) addRecipientIfNotPresented(domainRecipient notification.Recipient) error {
	if len(r.cache) >= r.storageLimit {
		return notification.ErrStorageLimitExceeded
	}

	for _, cacheRecipient := range r.cache {
		if cacheRecipient.TelegramID == domainRecipient.TelegramID {
			return notification.ErrAlreadyExists
		}
	}

	r.cache = append(r.cache, domainRecipient)

	return nil
}

func (r *recipientStorageFs) deleteRecipientIfPresented(domainRecipient notification.Recipient) error {
	for i, cacheRecipient := range r.cache {
		if cacheRecipient.TelegramID == domainRecipient.TelegramID {
			r.cache = append(r.cache[:i], r.cache[i+1:]...)

			return nil
		}
	}

	return notification.ErrNotExists
}

func (r *recipientStorageFs) writeCache() error {
	f, err := os.Create(r.storageFile)

	defer r.logger.CloseWithLog(f)

	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	encoder := json.NewEncoder(f)

	if err := encoder.Encode(r.fromDomain(r.cache)); err != nil {
		return fmt.Errorf("failed to write recipients: %w", err)
	}

	return nil
}

var errNoFileExists = fmt.Errorf("no file exists")

func (r *recipientStorageFs) readAllToCache() error {
	f, err := os.Open(r.storageFile)

	if os.IsNotExist(err) {
		return errNoFileExists
	}

	defer r.logger.CloseWithLog(f)

	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	decoder := json.NewDecoder(f)

	recipients := make([]recipient, 0)

	if err := decoder.Decode(&recipients); err != nil {
		return fmt.Errorf("failed to decode file: %w", err)
	}

	r.cache = r.toDomain(recipients)

	return nil
}

func (r *recipientStorageFs) createStorage() error {
	f, err := os.Create(r.storageFile)

	defer r.logger.CloseWithLog(f)

	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	if _, err := f.Write([]byte{'[', ']'}); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (r *recipientStorageFs) fromDomain(domainRecipients []notification.Recipient) []recipient {
	recipients := make([]recipient, 0, len(domainRecipients))

	for _, domainRecipient := range domainRecipients {
		recipients = append(recipients, recipient{
			TelegramID: domainRecipient.TelegramID,
			ID:         domainRecipient.ID,
			CD:         domainRecipient.CD,
		})
	}

	return recipients
}

func (r *recipientStorageFs) toDomain(recipients []recipient) []notification.Recipient {
	domainRecipients := make([]notification.Recipient, 0, len(recipients))

	for _, recipientObj := range recipients {
		domainRecipients = append(domainRecipients, notification.Recipient{
			TelegramID: recipientObj.TelegramID,
			ID:         recipientObj.ID,
			CD:         recipientObj.CD,
		})
	}

	return domainRecipients
}
