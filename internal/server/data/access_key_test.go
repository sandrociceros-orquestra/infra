package data

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/opt"

	"github.com/infrahq/infra/internal/generate"
	"github.com/infrahq/infra/internal/server/models"
	"github.com/infrahq/infra/uid"
)

func TestCreateAccessKey(t *testing.T) {
	runDBTests(t, func(t *testing.T, db *DB) {
		jerry := &models.Identity{Name: "jseinfeld@infrahq.com"}

		err := CreateIdentity(db, jerry)
		assert.NilError(t, err)

		infraProviderID := InfraProvider(db).ID

		t.Run("all default values", func(t *testing.T) {
			key := &models.AccessKey{
				IssuedFor:  jerry.ID,
				ProviderID: infraProviderID,
			}
			pair, err := CreateAccessKey(db, key)
			assert.NilError(t, err)

			expected := &models.AccessKey{
				Model: models.Model{
					ID:        uid.ID(12345),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				IssuedFor:      jerry.ID,
				ProviderID:     infraProviderID,
				KeyID:          "<any-string>",
				Secret:         "<any-string>",
				ExpiresAt:      time.Now().Add(12 * time.Hour),
				Name:           fmt.Sprintf("%s-%s", jerry.Name, key.ID.String()),
				SecretChecksum: secretChecksum(key.Secret),
			}
			assert.DeepEqual(t, key, expected, cmpAccessKey)
			assert.Equal(t, pair, key.KeyID+"."+key.Secret)

			// check that we can fetch the same value from the db
			fromDB, err := GetAccessKey(db, ByID(key.ID))
			assert.NilError(t, err)

			// fromDB should not have the secret value
			key.Secret = ""
			assert.DeepEqual(t, fromDB, key, cmpopts.EquateEmpty(), cmpTimeWithDBPrecision)
		})

		t.Run("all values", func(t *testing.T) {
			key := &models.AccessKey{
				Model: models.Model{
					ID:        uid.ID(512512),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Name:              "the-key",
				IssuedFor:         jerry.ID,
				ProviderID:        infraProviderID,
				ExpiresAt:         time.Now().Add(time.Hour),
				Extension:         3 * time.Hour,
				ExtensionDeadline: time.Now().Add(time.Minute),
				KeyID:             "0123456789",
				Secret:            "012345678901234567890123",
				Scopes:            []string{"first", "third"},
			}
			pair, err := CreateAccessKey(db, key)
			assert.NilError(t, err)
			assert.Equal(t, pair, key.KeyID+"."+key.Secret)

			// check that we can fetch the same value from the db
			fromDB, err := GetAccessKey(db, ByID(key.ID))
			assert.NilError(t, err)
			// fromDB should not have the secret value
			key.Secret = ""
			assert.DeepEqual(t, fromDB, key, cmpTimeWithDBPrecision)
		})

		t.Run("invalid specified key id length", func(t *testing.T) {
			key := &models.AccessKey{
				KeyID:      "too-short",
				IssuedFor:  jerry.ID,
				ProviderID: InfraProvider(db).ID,
			}
			_, err := CreateAccessKey(db, key)
			assert.Error(t, err, "invalid key length")
		})

		t.Run("invalid specified key secret length", func(t *testing.T) {
			key := &models.AccessKey{
				Secret:     "too-short",
				IssuedFor:  jerry.ID,
				ProviderID: InfraProvider(db).ID,
			}
			_, err := CreateAccessKey(db, key)
			assert.Error(t, err, "invalid secret length")
		})
	})
}

var cmpModel = cmp.Options{
	cmp.FilterPath(opt.PathField(models.Model{}, "ID"), anyValidUID),
	cmp.FilterPath(opt.PathField(models.Model{}, "CreatedAt"), opt.TimeWithThreshold(2*time.Second)),
	cmp.FilterPath(opt.PathField(models.Model{}, "UpdatedAt"), opt.TimeWithThreshold(2*time.Second)),
}

var cmpAccessKey = cmp.Options{
	cmpModel,
	cmp.FilterPath(opt.PathField(models.AccessKey{}, "KeyID"), nonZeroString),
	cmp.FilterPath(opt.PathField(models.AccessKey{}, "Secret"), nonZeroString),
	cmp.FilterPath(opt.PathField(models.AccessKey{}, "ExpiresAt"), opt.TimeWithThreshold(time.Second)),
}

var nonZeroString = cmp.Comparer(func(x, y string) bool {
	if x == "" || y == "" {
		return false
	}
	if x == "<any-string>" || y == "<any-string>" {
		return true
	}
	return false
})

var anyValidUID = cmp.Comparer(func(x, y uid.ID) bool {
	return x > 0 && y > 0
})

// PostgreSQL only has microsecond precision
var cmpTimeWithDBPrecision = cmpopts.EquateApproxTime(time.Microsecond)

func createAccessKeyWithExtensionDeadline(t *testing.T, db GormTxn, ttl, extensionDeadline time.Duration) (string, *models.AccessKey) {
	identity := &models.Identity{Name: "Wall-E"}
	err := CreateIdentity(db, identity)
	assert.NilError(t, err)

	token := &models.AccessKey{
		IssuedFor:         identity.ID,
		ProviderID:        InfraProvider(db).ID,
		ExpiresAt:         time.Now().Add(ttl),
		ExtensionDeadline: time.Now().Add(extensionDeadline).UTC(),
	}

	body, err := CreateAccessKey(db, token)
	assert.NilError(t, err)

	return body, token
}

func TestCheckAccessKeySecret(t *testing.T) {
	runDBTests(t, func(t *testing.T, db *DB) {
		body, _ := createTestAccessKey(t, db, time.Hour*5)

		_, err := ValidateAccessKey(db, body)
		assert.NilError(t, err)

		random := generate.MathRandom(models.AccessKeySecretLength, generate.CharsetAlphaNumeric)
		authorization := fmt.Sprintf("%s.%s", strings.Split(body, ".")[0], random)

		_, err = ValidateAccessKey(db, authorization)
		assert.Error(t, err, "access key invalid secret")
	})
}

func TestDeleteAccessKey(t *testing.T) {
	runDBTests(t, func(t *testing.T, db *DB) {
		_, token := createTestAccessKey(t, db, time.Minute*5)

		_, err := GetAccessKey(db, ByID(token.ID))
		assert.NilError(t, err)

		err = DeleteAccessKey(db, token.ID)
		assert.NilError(t, err)

		_, err = GetAccessKey(db, ByID(token.ID))
		assert.Error(t, err, "record not found")
	})
}

func TestDeleteAccessKeys(t *testing.T) {
	runDBTests(t, func(t *testing.T, db *DB) {
		provider := &models.Provider{Name: "azure", Kind: models.ProviderKindAzure}
		otherProvider := &models.Provider{Name: "other", Kind: models.ProviderKindGoogle}
		createProviders(t, db, provider, otherProvider)

		user := &models.Identity{Name: "main@example.com"}
		otherUser := &models.Identity{Name: "other@example.com"}
		createIdentities(t, db, user, otherUser)

		t.Run("empty options", func(t *testing.T) {
			err := DeleteAccessKeys(db, DeleteAccessKeysOptions{})
			assert.ErrorContains(t, err, "requires an ID to delete")
		})

		t.Run("by user id", func(t *testing.T) {
			tx := txnForTestCase(t, db)
			key1 := &models.AccessKey{IssuedFor: user.ID, ProviderID: provider.ID}
			key2 := &models.AccessKey{IssuedFor: user.ID, ProviderID: otherProvider.ID}
			toKeep := &models.AccessKey{IssuedFor: otherUser.ID, ProviderID: otherProvider.ID}
			createAccessKeys(t, tx, key1, key2, toKeep)

			err := DeleteAccessKeys(tx, DeleteAccessKeysOptions{ByUserID: user.ID})
			assert.NilError(t, err)

			remaining, err := ListAccessKeys(tx, nil)
			assert.NilError(t, err)
			expected := []models.AccessKey{
				{Model: models.Model{ID: toKeep.ID}},
			}
			assert.DeepEqual(t, remaining, expected, cmpModelByID)
		})

		t.Run("by provider id", func(t *testing.T) {
			tx := txnForTestCase(t, db)
			key1 := &models.AccessKey{IssuedFor: user.ID, ProviderID: provider.ID}
			key2 := &models.AccessKey{IssuedFor: otherUser.ID, ProviderID: provider.ID}
			toKeep := &models.AccessKey{IssuedFor: user.ID, ProviderID: otherProvider.ID}
			createAccessKeys(t, tx, key1, key2, toKeep)

			err := DeleteAccessKeys(tx, DeleteAccessKeysOptions{ByProviderID: provider.ID})
			assert.NilError(t, err)

			remaining, err := ListAccessKeys(tx, nil)
			assert.NilError(t, err)
			expected := []models.AccessKey{
				{Model: models.Model{ID: toKeep.ID}},
			}
			assert.DeepEqual(t, remaining, expected, cmpModelByID)
		})
	})
}

func createAccessKeys(t *testing.T, db GormTxn, keys ...*models.AccessKey) {
	t.Helper()
	for i := range keys {
		_, err := CreateAccessKey(db, keys[i])
		assert.NilError(t, err)
	}
}

type primaryKeyable interface {
	Primary() uid.ID
}

var cmpModelByID = cmp.Comparer(func(x, y primaryKeyable) bool {
	return x.Primary() == y.Primary()
})

func TestCheckAccessKeyExpired(t *testing.T) {
	runDBTests(t, func(t *testing.T, db *DB) {
		body, _ := createTestAccessKey(t, db, -1*time.Hour)

		_, err := ValidateAccessKey(db, body)
		assert.ErrorIs(t, err, ErrAccessKeyExpired)
	})
}

func TestCheckAccessKeyPastExtensionDeadline(t *testing.T) {
	runDBTests(t, func(t *testing.T, db *DB) {
		body, _ := createAccessKeyWithExtensionDeadline(t, db, 1*time.Hour, -1*time.Hour)

		_, err := ValidateAccessKey(db, body)
		assert.ErrorIs(t, err, ErrAccessKeyDeadlineExceeded)
	})
}

func TestListAccessKeys(t *testing.T) {
	runDBTests(t, func(t *testing.T, db *DB) {
		user := &models.Identity{Name: "tmp@infrahq.com"}
		err := CreateIdentity(db, user)
		assert.NilError(t, err)

		token := &models.AccessKey{
			Name:       "first",
			Model:      models.Model{ID: 0},
			IssuedFor:  user.ID,
			ProviderID: InfraProvider(db).ID,
			ExpiresAt:  time.Now().Add(time.Hour).UTC(),
			KeyID:      "1234567890",
		}
		_, err = CreateAccessKey(db, token)
		assert.NilError(t, err)

		token = &models.AccessKey{
			Name:       "second",
			Model:      models.Model{ID: 1},
			IssuedFor:  user.ID,
			ProviderID: InfraProvider(db).ID,
			ExpiresAt:  time.Now().Add(-time.Hour).UTC(),
			KeyID:      "1234567891",
		}
		_, err = CreateAccessKey(db, token)
		assert.NilError(t, err)

		token = &models.AccessKey{
			Name:              "third",
			Model:             models.Model{ID: 2},
			IssuedFor:         user.ID,
			ProviderID:        InfraProvider(db).ID,
			ExpiresAt:         time.Now().Add(time.Hour).UTC(),
			ExtensionDeadline: time.Now().Add(-time.Hour).UTC(),
			KeyID:             "1234567892",
		}
		_, err = CreateAccessKey(db, token)
		assert.NilError(t, err)

		keys, err := ListAccessKeys(db, nil, ByNotExpiredOrExtended())
		assert.NilError(t, err)
		assert.Equal(t, len(keys), 1)

		keys, err = ListAccessKeys(db, nil)
		assert.NilError(t, err)
		assert.Equal(t, len(keys), 3)
	})
}

func createTestAccessKey(t *testing.T, db GormTxn, sessionDuration time.Duration) (string, *models.AccessKey) {
	user := &models.Identity{Name: "tmp@infrahq.com"}
	err := CreateIdentity(db, user)
	assert.NilError(t, err)

	token := &models.AccessKey{
		IssuedFor:  user.ID,
		ProviderID: InfraProvider(db).ID,
		ExpiresAt:  time.Now().Add(sessionDuration),
	}

	body, err := CreateAccessKey(db, token)
	assert.NilError(t, err)

	return body, token
}
