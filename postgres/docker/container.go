package docker

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	tc "github.com/testcontainers/testcontainers-go"
)

const (
	defaultUser          = "db_user"
	defaultPassword      = "db_password"
	defaultPostgresImage = "postgres:15-alpine"
	defaultName          = "steron"
	dbPort               = "5432/tcp"
)

type PostgresContainer struct {
	tc.Container
	ctx      context.Context
	dbName   string
	user     string
	password string
	host     string
	port     int
}

func (p *PostgresContainer) Host() string {
	return p.host
}

func (p *PostgresContainer) Name() string {
	return p.dbName
}

func (p *PostgresContainer) User() string {
	return p.user
}

func (p *PostgresContainer) Port() int {
	return p.port
}

func (p *PostgresContainer) Password() string {
	return p.password
}

func (p *PostgresContainer) Cleanup() error {
	return p.Terminate(p.ctx)
}

func WithConfigFile(cfg string) tc.CustomizeRequestOption {
	return func(req *tc.GenericContainerRequest) {
		cfgFile := tc.ContainerFile{
			HostFilePath:      cfg,
			ContainerFilePath: "/etc/postgresql.conf",
			FileMode:          0o755,
		}

		req.Files = append(req.Files, cfgFile)
		req.Cmd = append(req.Cmd, "-c", "config_file=/etc/postgresql.conf")
	}
}

func WithDatabase(dbName string) tc.CustomizeRequestOption {
	return func(req *tc.GenericContainerRequest) {
		req.Env["POSTGRES_DB"] = dbName
	}
}

func WithInitScripts(scripts ...string) tc.CustomizeRequestOption {
	return func(req *tc.GenericContainerRequest) {
		initScripts := []tc.ContainerFile{}
		for _, script := range scripts {
			cf := tc.ContainerFile{
				HostFilePath:      script,
				ContainerFilePath: "/docker-entrypoint-initdb.d/" + filepath.Base(script),
				FileMode:          0o755,
			}
			initScripts = append(initScripts, cf)
		}
		req.Files = append(req.Files, initScripts...)
	}
}

func WithPassword(password string) tc.CustomizeRequestOption {
	return func(req *tc.GenericContainerRequest) {
		req.Env["POSTGRES_PASSWORD"] = password
	}
}

func WithUsername(user string) tc.CustomizeRequestOption {
	return func(req *tc.GenericContainerRequest) {
		if user == "" {
			user = defaultUser
		}

		req.Env["POSTGRES_USER"] = user
	}
}

func WithEnv(env map[string]string) tc.CustomizeRequestOption {
	return func(req *tc.GenericContainerRequest) {
		req.Env = env
	}
}

func WithEnvValue(key, value string) tc.CustomizeRequestOption {
	return func(req *tc.GenericContainerRequest) {
		req.Env[key] = value
	}
}

func WithImage(image string) tc.CustomizeRequestOption {
	return func(req *tc.GenericContainerRequest) {
		if len(image) == 0 {
			return
		}
		req.Image = image
	}
}

func WithExposedPorts(ports []string) tc.CustomizeRequestOption {
	return func(req *tc.GenericContainerRequest) {
		req.ExposedPorts = ports
	}
}

func WithEntrypoint(entrypoint []string) tc.CustomizeRequestOption {
	return func(req *tc.GenericContainerRequest) {
		req.Entrypoint = entrypoint
	}
}

func WithCmd(cmd []string) tc.CustomizeRequestOption {
	return func(req *tc.GenericContainerRequest) {
		req.Cmd = cmd
	}
}

func WithLifecycleHooks(lifecycleHooks []tc.ContainerLifecycleHooks) tc.CustomizeRequestOption {
	return func(req *tc.GenericContainerRequest) {
		req.LifecycleHooks = lifecycleHooks
	}
}

func RunContainer(opts ...tc.ContainerCustomizer) (*PostgresContainer, error) {
	ctx := context.Background()
	req := tc.ContainerRequest{
		Image: defaultPostgresImage,
		Env: map[string]string{
			"POSTGRES_USER":     defaultUser,
			"POSTGRES_PASSWORD": defaultPassword,
			"POSTGRES_DB":       defaultName,
		},
		ExposedPorts: []string{dbPort},
		Cmd:          []string{"postgres", "-c", "fsync=off"},
	}

	genericContainerReq := tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		opt.Customize(&genericContainerReq)
	}

	container, err := tc.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, err
	}

	user := req.Env["POSTGRES_USER"]
	password := req.Env["POSTGRES_PASSWORD"]
	dbName := req.Env["POSTGRES_DB"]

	pc := new(PostgresContainer)
	for i := 0; i < 3; i++ {
		pc, err = getPorts(ctx, container)
		if err == nil {
			break
		}
		time.Sleep(time.Millisecond * 50)
	}
	if err != nil {
		return nil, err
	}

	pg := &PostgresContainer{
		Container: container,
		dbName:    dbName,
		password:  password,
		user:      user,
		host:      pc.host,
		port:      pc.port,
	}

	return pg, nil
}

func getPorts(ctx context.Context, container tc.Container) (*PostgresContainer, error) {
	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("postgres get host error: %w", err)
	}
	ports, err := container.Ports(ctx)
	if err != nil {
		return nil, fmt.Errorf("postgres get port error: %w", err)
	}

	var hostPost string
	if len(ports[dbPort]) > 0 {
		hostPost = ports[dbPort][0].HostPort
	}

	port, err := strconv.Atoi(hostPost)
	if err != nil {
		return nil, fmt.Errorf("port '%s' parse error: %w", hostPost, err)
	}

	pg := &PostgresContainer{
		host: host,
		port: port,
	}

	return pg, nil
}
