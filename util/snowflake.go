package util

import "github.com/bwmarrin/snowflake"

// Generator is a global Snowflake node for process-wide ID generation.
var Generator *snowflake.Node

func init() {
	node, err := snowflake.NewNode(1)
	if err != nil {
		panic(err)
	}
	Generator = node
}

func NextSnowflakeID() int64 {
	return Generator.Generate().Int64()
}

func NextSnowflakeIDString() string {
	return Generator.Generate().String()
}
