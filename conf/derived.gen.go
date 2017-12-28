// Code generated by goderive DO NOT EDIT.

package conf

// deriveCloneBaseConfig returns a clone of the src parameter.
func deriveCloneBaseConfig(src BaseConfig) BaseConfig {
	dst := new(BaseConfig)
	deriveDeepCopy(dst, &src)
	return *dst
}

// deriveDeepCopy recursively copies the contents of src into dst.
func deriveDeepCopy(dst, src *BaseConfig) {
	if src.TcpSource == nil {
		dst.TcpSource = nil
	} else {
		if dst.TcpSource != nil {
			if len(src.TcpSource) > len(dst.TcpSource) {
				if cap(dst.TcpSource) >= len(src.TcpSource) {
					dst.TcpSource = (dst.TcpSource)[:len(src.TcpSource)]
				} else {
					dst.TcpSource = make([]TcpSourceConfig, len(src.TcpSource))
				}
			} else if len(src.TcpSource) < len(dst.TcpSource) {
				dst.TcpSource = (dst.TcpSource)[:len(src.TcpSource)]
			}
		} else {
			dst.TcpSource = make([]TcpSourceConfig, len(src.TcpSource))
		}
		deriveDeepCopy_(dst.TcpSource, src.TcpSource)
	}
	if src.UdpSource == nil {
		dst.UdpSource = nil
	} else {
		if dst.UdpSource != nil {
			if len(src.UdpSource) > len(dst.UdpSource) {
				if cap(dst.UdpSource) >= len(src.UdpSource) {
					dst.UdpSource = (dst.UdpSource)[:len(src.UdpSource)]
				} else {
					dst.UdpSource = make([]UdpSourceConfig, len(src.UdpSource))
				}
			} else if len(src.UdpSource) < len(dst.UdpSource) {
				dst.UdpSource = (dst.UdpSource)[:len(src.UdpSource)]
			}
		} else {
			dst.UdpSource = make([]UdpSourceConfig, len(src.UdpSource))
		}
		deriveDeepCopy_1(dst.UdpSource, src.UdpSource)
	}
	if src.RelpSource == nil {
		dst.RelpSource = nil
	} else {
		if dst.RelpSource != nil {
			if len(src.RelpSource) > len(dst.RelpSource) {
				if cap(dst.RelpSource) >= len(src.RelpSource) {
					dst.RelpSource = (dst.RelpSource)[:len(src.RelpSource)]
				} else {
					dst.RelpSource = make([]RelpSourceConfig, len(src.RelpSource))
				}
			} else if len(src.RelpSource) < len(dst.RelpSource) {
				dst.RelpSource = (dst.RelpSource)[:len(src.RelpSource)]
			}
		} else {
			dst.RelpSource = make([]RelpSourceConfig, len(src.RelpSource))
		}
		deriveDeepCopy_2(dst.RelpSource, src.RelpSource)
	}
	if src.KafkaSource == nil {
		dst.KafkaSource = nil
	} else {
		if dst.KafkaSource != nil {
			if len(src.KafkaSource) > len(dst.KafkaSource) {
				if cap(dst.KafkaSource) >= len(src.KafkaSource) {
					dst.KafkaSource = (dst.KafkaSource)[:len(src.KafkaSource)]
				} else {
					dst.KafkaSource = make([]KafkaSourceConfig, len(src.KafkaSource))
				}
			} else if len(src.KafkaSource) < len(dst.KafkaSource) {
				dst.KafkaSource = (dst.KafkaSource)[:len(src.KafkaSource)]
			}
		} else {
			dst.KafkaSource = make([]KafkaSourceConfig, len(src.KafkaSource))
		}
		deriveDeepCopy_3(dst.KafkaSource, src.KafkaSource)
	}
	if src.GraylogSource == nil {
		dst.GraylogSource = nil
	} else {
		if dst.GraylogSource != nil {
			if len(src.GraylogSource) > len(dst.GraylogSource) {
				if cap(dst.GraylogSource) >= len(src.GraylogSource) {
					dst.GraylogSource = (dst.GraylogSource)[:len(src.GraylogSource)]
				} else {
					dst.GraylogSource = make([]GraylogSourceConfig, len(src.GraylogSource))
				}
			} else if len(src.GraylogSource) < len(dst.GraylogSource) {
				dst.GraylogSource = (dst.GraylogSource)[:len(src.GraylogSource)]
			}
		} else {
			dst.GraylogSource = make([]GraylogSourceConfig, len(src.GraylogSource))
		}
		deriveDeepCopy_4(dst.GraylogSource, src.GraylogSource)
	}
	dst.Store = src.Store
	if src.Parsers == nil {
		dst.Parsers = nil
	} else {
		if dst.Parsers != nil {
			if len(src.Parsers) > len(dst.Parsers) {
				if cap(dst.Parsers) >= len(src.Parsers) {
					dst.Parsers = (dst.Parsers)[:len(src.Parsers)]
				} else {
					dst.Parsers = make([]ParserConfig, len(src.Parsers))
				}
			} else if len(src.Parsers) < len(dst.Parsers) {
				dst.Parsers = (dst.Parsers)[:len(src.Parsers)]
			}
		} else {
			dst.Parsers = make([]ParserConfig, len(src.Parsers))
		}
		copy(dst.Parsers, src.Parsers)
	}
	dst.Journald = src.Journald
	dst.Metrics = src.Metrics
	dst.Accounting = src.Accounting
	dst.Main = src.Main
	field := new(KafkaDestConfig)
	deriveDeepCopy_5(field, &src.KafkaDest)
	dst.KafkaDest = *field
	dst.UdpDest = src.UdpDest
	dst.TcpDest = src.TcpDest
	dst.HTTPDest = src.HTTPDest
	dst.RelpDest = src.RelpDest
	dst.FileDest = src.FileDest
	dst.StderrDest = src.StderrDest
	dst.GraylogDest = src.GraylogDest
}

// deriveDeepCopy_ recursively copies the contents of src into dst.
func deriveDeepCopy_(dst, src []TcpSourceConfig) {
	for src_i, src_value := range src {
		field := new(TcpSourceConfig)
		deriveDeepCopy_6(field, &src_value)
		dst[src_i] = *field
	}
}

// deriveDeepCopy_1 recursively copies the contents of src into dst.
func deriveDeepCopy_1(dst, src []UdpSourceConfig) {
	for src_i, src_value := range src {
		field := new(UdpSourceConfig)
		deriveDeepCopy_7(field, &src_value)
		dst[src_i] = *field
	}
}

// deriveDeepCopy_2 recursively copies the contents of src into dst.
func deriveDeepCopy_2(dst, src []RelpSourceConfig) {
	for src_i, src_value := range src {
		field := new(RelpSourceConfig)
		deriveDeepCopy_8(field, &src_value)
		dst[src_i] = *field
	}
}

// deriveDeepCopy_3 recursively copies the contents of src into dst.
func deriveDeepCopy_3(dst, src []KafkaSourceConfig) {
	for src_i, src_value := range src {
		field := new(KafkaSourceConfig)
		deriveDeepCopy_9(field, &src_value)
		dst[src_i] = *field
	}
}

// deriveDeepCopy_4 recursively copies the contents of src into dst.
func deriveDeepCopy_4(dst, src []GraylogSourceConfig) {
	for src_i, src_value := range src {
		field := new(GraylogSourceConfig)
		deriveDeepCopy_10(field, &src_value)
		dst[src_i] = *field
	}
}

// deriveDeepCopy_5 recursively copies the contents of src into dst.
func deriveDeepCopy_5(dst, src *KafkaDestConfig) {
	field := new(KafkaBaseConfig)
	deriveDeepCopy_11(field, &src.KafkaBaseConfig)
	dst.KafkaBaseConfig = *field
	dst.KafkaProducerBaseConfig = src.KafkaProducerBaseConfig
	dst.TlsBaseConfig = src.TlsBaseConfig
	dst.Insecure = src.Insecure
	dst.Format = src.Format
}

// deriveDeepCopy_6 recursively copies the contents of src into dst.
func deriveDeepCopy_6(dst, src *TcpSourceConfig) {
	field := new(SyslogSourceBaseConfig)
	deriveDeepCopy_12(field, &src.SyslogSourceBaseConfig)
	dst.SyslogSourceBaseConfig = *field
	dst.FilterSubConfig = src.FilterSubConfig
	dst.TlsBaseConfig = src.TlsBaseConfig
	dst.ClientAuthType = src.ClientAuthType
	dst.LineFraming = src.LineFraming
	dst.FrameDelimiter = src.FrameDelimiter
	dst.ConfID = src.ConfID
}

// deriveDeepCopy_7 recursively copies the contents of src into dst.
func deriveDeepCopy_7(dst, src *UdpSourceConfig) {
	field := new(SyslogSourceBaseConfig)
	deriveDeepCopy_12(field, &src.SyslogSourceBaseConfig)
	dst.SyslogSourceBaseConfig = *field
	dst.FilterSubConfig = src.FilterSubConfig
	dst.ConfID = src.ConfID
}

// deriveDeepCopy_8 recursively copies the contents of src into dst.
func deriveDeepCopy_8(dst, src *RelpSourceConfig) {
	field := new(SyslogSourceBaseConfig)
	deriveDeepCopy_12(field, &src.SyslogSourceBaseConfig)
	dst.SyslogSourceBaseConfig = *field
	dst.FilterSubConfig = src.FilterSubConfig
	dst.TlsBaseConfig = src.TlsBaseConfig
	dst.ClientAuthType = src.ClientAuthType
	dst.LineFraming = src.LineFraming
	dst.FrameDelimiter = src.FrameDelimiter
	dst.ConfID = src.ConfID
}

// deriveDeepCopy_9 recursively copies the contents of src into dst.
func deriveDeepCopy_9(dst, src *KafkaSourceConfig) {
	field := new(KafkaBaseConfig)
	deriveDeepCopy_11(field, &src.KafkaBaseConfig)
	dst.KafkaBaseConfig = *field
	dst.KafkaConsumerBaseConfig = src.KafkaConsumerBaseConfig
	dst.FilterSubConfig = src.FilterSubConfig
	dst.TlsBaseConfig = src.TlsBaseConfig
	dst.Insecure = src.Insecure
	dst.Format = src.Format
	dst.Encoding = src.Encoding
	dst.ConfID = src.ConfID
	dst.SessionTimeout = src.SessionTimeout
	dst.HeartbeatInterval = src.HeartbeatInterval
	dst.OffsetsMaxRetry = src.OffsetsMaxRetry
	dst.GroupID = src.GroupID
	if src.Topics == nil {
		dst.Topics = nil
	} else {
		if dst.Topics != nil {
			if len(src.Topics) > len(dst.Topics) {
				if cap(dst.Topics) >= len(src.Topics) {
					dst.Topics = (dst.Topics)[:len(src.Topics)]
				} else {
					dst.Topics = make([]string, len(src.Topics))
				}
			} else if len(src.Topics) < len(dst.Topics) {
				dst.Topics = (dst.Topics)[:len(src.Topics)]
			}
		} else {
			dst.Topics = make([]string, len(src.Topics))
		}
		copy(dst.Topics, src.Topics)
	}
}

// deriveDeepCopy_10 recursively copies the contents of src into dst.
func deriveDeepCopy_10(dst, src *GraylogSourceConfig) {
	field := new(SyslogSourceBaseConfig)
	deriveDeepCopy_12(field, &src.SyslogSourceBaseConfig)
	dst.SyslogSourceBaseConfig = *field
	dst.FilterSubConfig = src.FilterSubConfig
	dst.ConfID = src.ConfID
}

// deriveDeepCopy_11 recursively copies the contents of src into dst.
func deriveDeepCopy_11(dst, src *KafkaBaseConfig) {
	if src.Brokers == nil {
		dst.Brokers = nil
	} else {
		if dst.Brokers != nil {
			if len(src.Brokers) > len(dst.Brokers) {
				if cap(dst.Brokers) >= len(src.Brokers) {
					dst.Brokers = (dst.Brokers)[:len(src.Brokers)]
				} else {
					dst.Brokers = make([]string, len(src.Brokers))
				}
			} else if len(src.Brokers) < len(dst.Brokers) {
				dst.Brokers = (dst.Brokers)[:len(src.Brokers)]
			}
		} else {
			dst.Brokers = make([]string, len(src.Brokers))
		}
		copy(dst.Brokers, src.Brokers)
	}
	dst.ClientID = src.ClientID
	dst.Version = src.Version
	dst.ChannelBufferSize = src.ChannelBufferSize
	dst.MaxOpenRequests = src.MaxOpenRequests
	dst.DialTimeout = src.DialTimeout
	dst.ReadTimeout = src.ReadTimeout
	dst.WriteTimeout = src.WriteTimeout
	dst.KeepAlive = src.KeepAlive
	dst.MetadataRetryMax = src.MetadataRetryMax
	dst.MetadataRetryBackoff = src.MetadataRetryBackoff
	dst.MetadataRefreshFrequency = src.MetadataRefreshFrequency
}

// deriveDeepCopy_12 recursively copies the contents of src into dst.
func deriveDeepCopy_12(dst, src *SyslogSourceBaseConfig) {
	if src.Ports == nil {
		dst.Ports = nil
	} else {
		if dst.Ports != nil {
			if len(src.Ports) > len(dst.Ports) {
				if cap(dst.Ports) >= len(src.Ports) {
					dst.Ports = (dst.Ports)[:len(src.Ports)]
				} else {
					dst.Ports = make([]int, len(src.Ports))
				}
			} else if len(src.Ports) < len(dst.Ports) {
				dst.Ports = (dst.Ports)[:len(src.Ports)]
			}
		} else {
			dst.Ports = make([]int, len(src.Ports))
		}
		copy(dst.Ports, src.Ports)
	}
	dst.BindAddr = src.BindAddr
	dst.UnixSocketPath = src.UnixSocketPath
	dst.Format = src.Format
	dst.DontParseSD = src.DontParseSD
	dst.KeepAlive = src.KeepAlive
	dst.KeepAlivePeriod = src.KeepAlivePeriod
	dst.Timeout = src.Timeout
	dst.Encoding = src.Encoding
}
