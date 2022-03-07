package db

import (
	"database/sql"

	"github.com/aragon/zkmultisig-node/types"
)

// SQLite represents the SQLite database
type SQLite struct {
	db *sql.DB
}

// NewSQLite returns a new *SQLite database
func NewSQLite(db *sql.DB) *SQLite {
	return &SQLite{
		db: db,
	}
}

// Migrate creates the tables needed for the database
func (r *SQLite) Migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS votepackages(
		censusRoot BLOB NOT NULL,
		signature BLOB NOT NULL,
		indx INTEGER NOT NULL PRIMARY KEY UNIQUE,
		publicKey BLOB NOT NULL UNIQUE,
		merkleproof BLOB NOT NULL UNIQUE,
		vote BLOB NOT NULL,
		insertedDatetime DATETIME
	);
	`

	_, err := r.db.Exec(query)
	return err
}

// StoreVotePackage stores the given types.VotePackage for the given CensusRoot
func (r *SQLite) StoreVotePackage(censusRoot []byte, vote types.VotePackage) error {
	sqlAddvote := `
	INSERT INTO votepackages(
		censusRoot,
		signature,
		indx,
		publicKey,
		merkleproof,
		vote,
		insertedDatetime
	) values(?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`

	stmt, err := r.db.Prepare(sqlAddvote)
	if err != nil {
		return err
	}
	defer stmt.Close() //nolint:errcheck

	_, err = stmt.Exec(censusRoot, vote.Signature[:],
		vote.CensusProof.Index, vote.CensusProof.PublicKey,
		vote.CensusProof.MerkleProof, vote.Vote)
	if err != nil {
		return err
	}
	return nil
}

// ReadVotePackagesByCensusRoot reads all the stored types.VotePackage for the
// given CensusRoot
func (r *SQLite) ReadVotePackagesByCensusRoot(censusRoot []byte) ([]types.VotePackage, error) {
	// TODO add pagination
	sqlReadall := `
	SELECT signature, indx, publicKey, merkleproof, vote FROM votepackages
	WHERE censusRoot = ?
	ORDER BY datetime(InsertedDatetime) DESC
	`

	rows, err := r.db.Query(sqlReadall, censusRoot)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var votes []types.VotePackage
	for rows.Next() {
		vote := types.VotePackage{}
		var sigBytes []byte
		err = rows.Scan(&sigBytes, &vote.CensusProof.Index,
			&vote.CensusProof.PublicKey, &vote.CensusProof.MerkleProof,
			&vote.Vote)
		if err != nil {
			return nil, err
		}
		copy(vote.Signature[:], sigBytes)
		votes = append(votes, vote)
	}
	return votes, nil
}

// func (r *SQLite) ReadVoteByPublicKeyAndCensusRoot(censusRoot []byte) (
// 	[]types.VotePackage, error) {
// func (r *SQLite) ReadVotesByPublicKey(censusRoot []byte) ([]types.VotePackage, error) {