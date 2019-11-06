package db

import (
	"bitbucket.org/zlacki/rscgo/pkg/server/crypto"
	"strconv"

	"bitbucket.org/zlacki/rscgo/pkg/server/config"
	"bitbucket.org/zlacki/rscgo/pkg/server/errors"
	"bitbucket.org/zlacki/rscgo/pkg/server/log"
	"bitbucket.org/zlacki/rscgo/pkg/server/world"
	"bitbucket.org/zlacki/rscgo/pkg/strutil"
)

// FIXME: This is an exact copy from the old database code model to its own, isolated package, and the result is clumsy and needs to be tidied up badly.

//CreatePlayer Creates a new entry in the player SQLite3 database with the specified credentials.
// Returns true if successful, otherwise returns false.
func CreatePlayer(username, password string) bool {
	database := Open(config.PlayerDB())
	defer database.Close()

	tx, err := database.Begin()
	if err != nil {
		log.Info.Println("CreatePlayer(): Could not begin transaction for new player.")
		return false
	}

	s, err := tx.Exec("INSERT INTO player2(username, userhash, password, x, y, group_id) VALUES(?, ?, ?, 220, 445, 0)", username, strutil.Base37.Encode(username), crypto.Hash(password))
	if err != nil {
		log.Info.Println("CreatePlayer(): Could not insert new player profile information:", err)
		return false
	}
	playerID, err := s.LastInsertId()
	if err != nil || playerID < 0 {
		log.Info.Printf("CreatePlayer(): Could not retrieve player database ID(got %d):\n%v", playerID, err)
		return false
	}
	_, err = tx.Exec("INSERT INTO appearance VALUES(?, 2, 8, 14, 0, 1, 2)", playerID)
	if err != nil {
		log.Info.Println("CreatePlayer(): Could not insert new player profile information:", err)
		return false
	}
	if err := tx.Commit(); err != nil {
		log.Warning.Println("CreatePlayer(): Error committing transaction for new player:", err)
		return false
	}

	return true
}

//UsernameExists Returns true if there is a player with the name 'username' in the player database, otherwise returns false.
func UsernameExists(username string) bool {
	database := Open(config.PlayerDB())
	defer database.Close()
	s, err := database.Query("SELECT id FROM player2 WHERE userhash=?", strutil.Base37.Encode(username))
	defer s.Close()
	if err != nil {
		log.Info.Println("UsernameTaken: Could not query player profile information:", err)
		// return true just to be safe since we could not check
		return true
	}
	return s.Next()
}

//ValidCredentials Returns true if it finds a user with this username hash and password in the database, otherwise returns false
func ValidCredentials(userHash uint64, password string) bool {
	database := Open(config.PlayerDB())
	defer database.Close()
	rows, err := database.Query("SELECT id FROM player2 WHERE userhash=? AND password=?", userHash, password)
	defer rows.Close()
	if err != nil {
		log.Info.Println("Validate: Could not validate user credentials:", err)
		return false
	}
	return rows.Next()
}

//LoadPlayerProfile Looks for a player with the specified credentials in the player database.  Returns nil if it finds the player, otherwise returns an error.
func LoadPlayerProfile(usernameHash uint64, password string, loginReply chan byte, player *world.Player) error {
	database := Open(config.PlayerDB())
	defer database.Close()
	rows, err := database.Query("SELECT player.id, player.x, player.y, player.group_id, appearance.haircolour, appearance.topcolour, appearance.trousercolour, appearance.skincolour, appearance.head, appearance.body FROM player2 AS player INNER JOIN appearance AS appearance WHERE appearance.playerid=player.id AND player.userhash=? AND player.password=?", usernameHash, crypto.Hash(password))
	defer rows.Close()
	if err != nil {
		log.Info.Println("ValidatePlayer(uint64,string): Could not prepare query statement for player:", err)
		loginReply <- byte(3)
		return errors.NewDatabaseError(err.Error())
	}
	if !rows.Next() {
		loginReply <- byte(3)
		return errors.NewDatabaseError("Could not find player")
	}
	var x, y uint32
	rows.Scan(&player.DatabaseIndex, &x, &y, &player.Rank, &player.Appearance.HeadColor, &player.Appearance.BodyColor, &player.Appearance.LegsColor, &player.Appearance.SkinColor, &player.Appearance.Head, &player.Appearance.Body)
	player.X.Store(x)
	player.Y.Store(y)
	return nil
}

//UpdatePassword Updates the players password to password in the database.
func UpdatePassword(userHash uint64, password string) bool {
	database := Open(config.PlayerDB())
	defer database.Close()
	s, err := database.Exec("UPDATE player2 SET password=? WHERE userhash=?", password, userHash)
	if err != nil {
		log.Info.Println("UpdatePassword: Could not update player password:", err)
		return false
	}
	count, err := s.RowsAffected()
	if count <= 0 || err != nil {
		log.Info.Println("UpdatePassword: Could not update player password:", err)
		return false
	}
	return true
}

//HasRecoveryQuestions Returns true if this username has recovery questions assigned to it, otherwise returns false.
func HasRecoveryQuestions(userHash uint64) bool {
	database := Open(config.PlayerDB())
	defer database.Close()
	rows, err := database.Query("SELECT question1 FROM recovery_questions WHERE userhash=?", userHash)
	defer rows.Close()
	if err != nil {
		log.Info.Println("HasRecoveryQuestions: Could not search for recovery questions:", err)
		return false
	}
	return rows.Next()
}

//GetRecoveryQuestions Retrieves the recovery questions assigned to this username if any, otherwise returns nil
func GetRecoveryQuestions(userHash uint64) []string {
	database := Open(config.PlayerDB())
	defer database.Close()
	rows, err := database.Query("SELECT question1, question2, question3, question4, question5 FROM recovery_questions WHERE userhash=?", userHash)
	defer rows.Close()
	if err != nil {
		log.Info.Println("GetRecoveryQuestions: Could not find recovery questions:", err)
		return nil
	}

	var question1, question2, question3, question4, question5 string
	if rows.Next() {
		err := rows.Scan(&question1, &question2, &question3, &question4, &question5)
		if err != nil {
			log.Info.Println("GetRecoveryQuestions: Could not scan recovery questions to variables:", err)
			return nil
		}
		return []string{question1, question2, question3, question4, question5}
	}

	return nil
}

//SaveRecoveryQuestions Saves new recovery questions to the database.
func SaveRecoveryQuestions(userHash uint64, questions []string, answers []uint64) {

}

//LoadPlayerAttributes Looks for a player with the specified credentials in the player database.  Returns nil if it finds the player, otherwise returns an error.
func LoadPlayerAttributes(player *world.Player) error {
	database := Open(config.PlayerDB())
	defer database.Close()

	rows, err := database.Query("SELECT name, value FROM player_attr WHERE player_id=?", player.DatabaseIndex)
	defer rows.Close()
	if err != nil {
		log.Info.Println("LoadPlayer(uint64,string): Could not execute query statement for player attributes:", err)
		return errors.NewDatabaseError("Statement could not execute.")
	}
	for rows.Next() {
		var name, value string
		rows.Scan(&name, &value)
		switch value[0] {
		case 'i':
			val, err := strconv.ParseInt(value[1:], 10, 64)
			if err != nil {
				log.Info.Printf("Error loading int attribute[%v]: value=%v\n", name, value[1:])
				log.Info.Println(err)
			}
			player.Attributes.SetVar(name, int(val))
			break
		case 'l':
			val, err := strconv.ParseUint(value[1:], 10, 64)
			if err != nil {
				log.Info.Printf("Error loading long int attribute[%v]: value=%v\n", name, value[1:])
				log.Info.Println(err)
			}
			player.Attributes.SetVar(name, uint(val))
			break
		case 'b':
			val, err := strconv.ParseBool(value[1:])
			if err != nil {
				log.Info.Printf("Error loading boolean attribute[%v]: value=%v\n", name, value[1:])
				log.Info.Println(err)
			}
			player.Attributes.SetVar(name, val)
			break
		}
	}
	return nil
}

//LoadPlayerContacts Looks for a player with the specified credentials in the player database.  Returns nil if it finds the player, otherwise returns an error.
func LoadPlayerContacts(listType string, player *world.Player) error {
	database := Open(config.PlayerDB())
	defer database.Close()

	rows, err := database.Query("SELECT playerhash FROM playerlist WHERE playerid=? AND `type`=?", player.DatabaseIndex, listType)
	defer rows.Close()
	if err != nil {
		log.Info.Println("LoadPlayer(uint64,string): Could not execute query statement for player friends:", err)
		return errors.NewDatabaseError("Statement could not execute.")
	}
	for rows.Next() {
		var hash uint64
		rows.Scan(&hash)
		if listType == "friend" {
			player.FriendList[hash] = false
		} else {
			player.IgnoreList = append(player.IgnoreList, hash)
		}
	}
	return nil
}

//LoadPlayerInventory Looks for a player with the specified credentials in the player database.  Returns nil if it finds the player, otherwise returns an error.
func LoadPlayerInventory(player *world.Player) error {
	database := Open(config.PlayerDB())
	defer database.Close()
	rows, err := database.Query("SELECT itemid, amount, position, wielded FROM inventory WHERE playerid=?", player.DatabaseIndex)
	defer rows.Close()
	if err != nil {
		log.Info.Println("LoadPlayer(uint64,string): Could not execute query statement for player inventory:", err)
		return errors.NewDatabaseError("Statement could not execute.")
	}
	for rows.Next() {
		var id, amt, index int
		wielded := false
		rows.Scan(&id, &amt, &index, &wielded)
		index = player.Items.Put(id, amt)
		if e := GetEquipmentDefinition(id); e != nil && wielded {
			player.Items.Get(index).Worn = true
			player.Equips[e.Position] = e.Sprite
			player.SetAimPoints(player.AimPoints() + e.Aim)
			player.SetPowerPoints(player.PowerPoints() + e.Power)
			player.SetArmourPoints(player.ArmourPoints() + e.Armour)
			player.SetMagicPoints(player.MagicPoints() + e.Magic)
			player.SetPrayerPoints(player.PrayerPoints() + e.Prayer)
			player.SetRangedPoints(player.RangedPoints() + e.Ranged)
		}
	}
	return nil
}

//LoadPlayer Loads a player from the SQLite3 database, returns a login response code.
func LoadPlayer(player *world.Player, usernameHash uint64, password string, loginReply chan byte) {
	// If this fails, then the login information was incorrect, and we don't need to do anything else
	if err := LoadPlayerProfile(usernameHash, password, loginReply, player); err != nil {
		return
	}
	if err := LoadPlayerAttributes(player); err != nil {
		return
	}
	if err := LoadPlayerContacts("friend", player); err != nil {
		return
	}
	if err := LoadPlayerContacts("ignore", player); err != nil {
		return
	}
	if err := LoadPlayerInventory(player); err != nil {
		return
	}

	player.UserBase37 = usernameHash
	player.Username = strutil.Base37.Decode(usernameHash)
	if player.Rank == 2 {
		// Administrator
		loginReply <- 25
		return
	}
	if player.Rank == 1 {
		// Moderator
		loginReply <- 24
		return
	}
	if player.Reconnecting() {
		loginReply <- 1
		return
	}
	loginReply <- 0
	return
}

//SavePlayer Saves a player to the SQLite3 database.
func SavePlayer(player *world.Player) {
	db := Open(config.PlayerDB())
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		log.Info.Println("Save(): Could not begin transcaction for player update.")
		return
	}
	saveLocation := func() {
		rs, err := tx.Exec("UPDATE player2 SET x=?, y=? WHERE id=?", player.X.Load(), player.Y.Load(), player.DatabaseIndex)
		count, err := rs.RowsAffected()
		if err != nil {
			log.Warning.Println("Save(): UPDATE failed for player location:", err)
			if err := tx.Rollback(); err != nil {
				log.Warning.Println("Save(): Transaction location rollback failed:", err)
			}
			return
		}

		if count <= 0 {
			log.Info.Println("Save(): Affected nothing for location update!")
		}
	}
	saveAppearance := func() {
		// TODO: Should this just be attributes too??  Is that abusing the attributes table?
		appearance := player.Appearance
		rs, _ := tx.Exec("UPDATE appearance SET haircolour=?, topcolour=?, trousercolour=?, skincolour=?, head=?, body=? WHERE playerid=?", appearance.HeadColor, appearance.BodyColor, appearance.LegsColor, appearance.SkinColor, appearance.Head, appearance.Body, player.DatabaseIndex)
		count, err := rs.RowsAffected()
		if err != nil {
			log.Warning.Println("Save(): UPDATE failed for player appearance:", err)
			if err := tx.Rollback(); err != nil {
				log.Warning.Println("Save(): Transaction appearance rollback failed:", err)
			}
			return
		}

		if count <= 0 {
			log.Info.Println("Save(): Affected nothing for appearance update!")
		}
	}
	clearAttributes := func() {
		if _, err := tx.Exec("DELETE FROM player_attr WHERE player_id=?", player.DatabaseIndex); err != nil {
			log.Warning.Println("Save(): DELETE failed for player attribute:", err)
			if err := tx.Rollback(); err != nil {
				log.Warning.Println("Save(): Transaction delete attributes rollback failed:", err)
			}
			return
		}
	}
	insertAttribute := func(name string, value interface{}) {
		var val string
		switch value.(type) {
		case int:
			val = "i" + strconv.FormatInt(int64(value.(int)), 10)
		case uint:
			val = "l" + strconv.FormatUint(uint64(value.(uint)), 10)
		case bool:
			if v, ok := value.(bool); v && ok {
				val = "b1"
			} else {
				val = "b0"
			}
		}
		rs, _ := tx.Exec("INSERT INTO player_attr(player_id, name, value) VALUES(?, ?, ?)", player.DatabaseIndex, name, val)
		count, err := rs.RowsAffected()
		if err != nil {
			log.Warning.Println("Save(): INSERT failed for player attribute:", err)
			if err := tx.Rollback(); err != nil {
				log.Warning.Println("Save(): Transaction insert attribute rollback failed:", err)
			}
			return
		}

		if count <= 0 {
			log.Info.Println("Save(): Affected nothing for attribute insertion!")
		}
	}
	clearContactList := func(contactType string) {
		if _, err := tx.Exec("DELETE FROM playerlist WHERE playerid=? AND type=?", player.DatabaseIndex, contactType); err != nil {
			log.Warning.Println("Save(): DELETE failed for player friends:", err)
			if err := tx.Rollback(); err != nil {
				log.Warning.Println("Save(): Transaction delete friends rollback failed:", err)
			}
			return
		}
	}
	insertContactList := func(contactType string, hash uint64) {
		rs, _ := tx.Exec("INSERT INTO playerlist(playerid, playerhash, type) VALUES(?, ?, ?)", player.DatabaseIndex, hash, contactType)
		count, err := rs.RowsAffected()
		if err != nil {
			log.Warning.Println("Save(): INSERT failed for player friends:", err)
			if err := tx.Rollback(); err != nil {
				log.Warning.Println("Save(): Transaction insert friend rollback failed:", err)
			}
			return
		}

		if count <= 0 {
			log.Info.Println("Save(): Affected nothing for friend insertion!")
		}
	}
	clearItems := func() {
		if _, err := tx.Exec("DELETE FROM inventory WHERE playerid=?", player.DatabaseIndex); err != nil {
			log.Warning.Println("Save(): DELETE failed for player inventory:", err)
			if err := tx.Rollback(); err != nil {
				log.Warning.Println("Save(): Transaction delete inventory rollback failed:", err)
			}
			return
		}
	}
	insertItem := func(id, amt, index int, worn bool) {
		rs, _ := tx.Exec("INSERT INTO inventory(playerid, itemid, amount, position, wielded) VALUES(?, ?, ?, ?, ?)", player.DatabaseIndex, id, amt, index, worn)
		count, err := rs.RowsAffected()
		if err != nil {
			log.Warning.Println("Save(): INSERT failed for player items:", err)
			if err := tx.Rollback(); err != nil {
				log.Warning.Println("Save(): Transaction insert item rollback failed:", err)
			}
			return
		}

		if count <= 0 {
			log.Info.Println("Save(): Affected nothing for item insertion!")
		}
	}
	saveLocation()
	saveAppearance()
	clearAttributes()
	player.Attributes.Range(insertAttribute)
	clearContactList("friend")
	clearContactList("ignore")
	for hash := range player.FriendList {
		insertContactList("friend", hash)
	}
	for _, hash := range player.IgnoreList {
		insertContactList("ignore", hash)
	}
	clearItems()
	for _, item := range player.Items.List {
		insertItem(item.ID, item.Amount, item.Index, item.Worn)
	}

	if err := tx.Commit(); err != nil {
		log.Warning.Println("Save(): Error committing transaction for player update:", err)
	}
}
