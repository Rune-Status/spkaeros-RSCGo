package config

//TomlConfig A data structure representing the RSCGo TOML configuration file.
var TomlConfig struct {
	DataDir           string `toml:"data_directory"`
	Version           int    `toml:"version"`
	Port              int    `toml:"port"`
	MaxPlayers        int    `toml:"max_players"`
	PacketHandlerFile string `toml:"packet_handler_table"`
	Database          struct {
		PlayerDB string `toml:"player_db"`
		WorldDB  string `toml:"world_db"`
	} `toml:"database"`
	Crypto struct {
		RsaKeyFile     string `toml:"rsa_key"`
		HashSalt       string `toml:"hash_salt"`
		HashComplexity int    `toml:"hash_complexity"`
		HashMemory     int    `toml:"hash_memory"`
		HashLength     int    `toml:"hash_length"`
	} `toml:"crypto"`
}

//Verbosity Represents the level of verbosity with which the game should output debug information.
var Verbosity = int(0)

//Verbose Returns true if the current Verbosity level is above 0, otherwise returns false.
func Verbose() bool {
	return Verbosity > 0
}

//Port Returns the primary TCP/IP port to listen for incoming connections on
func Port() int {
	return TomlConfig.Port
}

//WSPort Returns the primary TCP/IP port to listen for incoming websocket connections on
func WSPort() int {
	return TomlConfig.Port + 1
}

func MaxPlayers() int {
	return TomlConfig.MaxPlayers
}

func DataDir() string {
	return TomlConfig.DataDir
}

func Version() int {
	return TomlConfig.Version
}

func PacketHandlers() string {
	return TomlConfig.PacketHandlerFile
}

func HashLength() int {
	return TomlConfig.Crypto.HashLength
}

func HashComplexity() int {
	return TomlConfig.Crypto.HashComplexity
}

func HashMemory() int {
	return TomlConfig.Crypto.HashMemory
}

func HashSalt() string {
	return TomlConfig.Crypto.HashSalt
}

func WorldDB() string {
	return TomlConfig.Database.WorldDB
}

func PlayerDB() string {
	return TomlConfig.Database.PlayerDB
}

func init() {
	TomlConfig.MaxPlayers = 1250
	TomlConfig.DataDir = "./data/"
	TomlConfig.Database.PlayerDB = "players.db"
	TomlConfig.Database.WorldDB = "world.db"
	TomlConfig.PacketHandlerFile = "packets.toml"
	TomlConfig.Crypto.HashComplexity = 15
	TomlConfig.Crypto.HashLength = 32
	TomlConfig.Crypto.HashMemory = 8
	TomlConfig.Crypto.HashSalt = "rscgo./GOLANG!RULES/.1994"
	TomlConfig.Version = 204
	TomlConfig.Port = 43594                  // = 43595 for websocket connections
	//TomlConfig.Crypto.RsaKeyFile = "rsa.der"
}
