/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/

package cmd

type KafkaCreds struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type DistributionConfig struct {
	KafkaCreds KafkaCreds `mapstructure:"kafka"`
}
