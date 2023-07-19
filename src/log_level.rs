use clap::ValueEnum;
use serde::{Deserialize, Serialize};
use tracing_subscriber::filter::LevelFilter;

#[derive(Default, Debug, Clone, PartialEq, Eq, ValueEnum, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum LogLevel {
	Trace,
	Debug,
	#[default]
	Info,
	Warn,
	Error,
}

impl LogLevel {
	pub fn to_tracing(&self) -> LevelFilter {
		match self {
			Self::Trace => LevelFilter::TRACE,
			Self::Debug => LevelFilter::DEBUG,
			Self::Info => LevelFilter::INFO,
			Self::Warn => LevelFilter::WARN,
			Self::Error => LevelFilter::ERROR,
		}
	}
}

impl std::fmt::Display for LogLevel {
	fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
		let s = match self {
			Self::Trace => "trace",
			Self::Debug => "debug",
			Self::Info => "info",
			Self::Warn => "warn",
			Self::Error => "error",
		};
		s.fmt(f)
	}
}

impl std::str::FromStr for LogLevel {
	type Err = String;

	fn from_str(s: &str) -> Result<Self, Self::Err> {
		match s {
			"trace" => Ok(Self::Trace),
			"debug" => Ok(Self::Debug),
			"info" => Ok(Self::Info),
			"warn" => Ok(Self::Warn),
			"error" => Ok(Self::Error),
			_ => Err(format!("Unknown log level: {s}")),
		}
	}
}