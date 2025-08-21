package configs

import "time"

// cfgStorage implement
// postgres
func (s PsqlStore) Host() string {
	return s.HostF
}
func (s PsqlStore) Port() string {
	return s.PortF
}
func (s PsqlStore) User() string {
	return s.UserF
}
func (s PsqlStore) Password() string {
	return s.PasswordF
}
func (s PsqlStore) Name() string {
	return s.NameF
}
func (s PsqlStore) SSLmode() string {
	return s.SSLmodeF
}

// redis
func (s RedisStore) Host() string {
	return s.HostF
}
func (s RedisStore) Port() string {
	return s.PortF
}
func (s RedisStore) DBNumber() int {
	return s.DBNumberF
}
func (s RedisStore) Password() string {
	return s.PasswordF
}

// cfgServerImplement
func (s Server) Address() string {
	return s.AddressF
}
func (s Server) Port() string {
	return s.PortF
}
func (s Server) ReadTimeout() time.Duration {
	return s.ReadTimeoutF
}
func (s Server) WriteTimeout() time.Duration {
	return s.WriteTimeoutF
}
func (s Server) ReadHeaderTimeout() time.Duration {
	return s.ReadHeaderTimeoutF
}
func (s Server) IdleTimeout() time.Duration {
	return s.IdleTimeoutF
}
