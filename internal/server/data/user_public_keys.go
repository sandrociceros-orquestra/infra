package data

import (
	"fmt"

	"github.com/infrahq/infra/internal/server/data/querybuilder"
	"github.com/infrahq/infra/internal/server/models"
	"github.com/infrahq/infra/uid"
)

type userPublicKeysTable models.UserPublicKey

func (t userPublicKeysTable) Table() string {
	return "user_public_keys"
}

func (t userPublicKeysTable) Columns() []string {
	return []string{"created_at", "deleted_at", "expires_at", "fingerprint", "id", "key_type", "name", "public_key", "updated_at", "user_id"}
}

func (t userPublicKeysTable) Values() []any {
	return []any{t.CreatedAt, t.DeletedAt, t.ExpiresAt, t.Fingerprint, t.ID, t.KeyType, t.Name, t.PublicKey, t.UpdatedAt, t.UserID}
}

func (t *userPublicKeysTable) ScanFields() []any {
	return []any{&t.CreatedAt, &t.DeletedAt, &t.ExpiresAt, &t.Fingerprint, &t.ID, &t.KeyType, &t.Name, &t.PublicKey, &t.UpdatedAt, &t.UserID}
}

func userPublicKeys(tx ReadTxn, userID uid.ID) ([]models.UserPublicKey, error) {
	table := userPublicKeysTable{}
	query := querybuilder.New("SELECT")
	query.B(columnsForSelect(table))
	query.B("FROM user_public_keys")
	query.B("WHERE deleted_at is null")
	query.B("AND user_id = ?", userID)

	rows, err := tx.Query(query.String(), query.Args...)
	if err != nil {
		return nil, err
	}
	return scanRows(rows, func(k *models.UserPublicKey) []any {
		return (*userPublicKeysTable)(k).ScanFields()
	})
}

func AddUserPublicKey(tx WriteTxn, key *models.UserPublicKey) error {
	switch {
	case key.UserID == 0:
		return fmt.Errorf("a userID is required")
	case key.Fingerprint == "":
		return fmt.Errorf("fingerprint is required")
	case key.KeyType == "":
		return fmt.Errorf("key type is required")
	}
	if err := insert(tx, (*userPublicKeysTable)(key)); err != nil {
		return err
	}

	// Create an ssh username for anyone that existed before we added the field
	// TODO: we have the user in the caller, can we use that instead of querying again?
	user, err := GetIdentity(tx, GetIdentityOptions{ByID: key.UserID})
	if err != nil {
		return err
	}
	if user.SSHUsername != "" {
		return nil
	}
	_, err = SetSSHUsername(tx, user)
	return err
}
