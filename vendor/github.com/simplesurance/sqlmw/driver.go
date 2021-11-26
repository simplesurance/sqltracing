package sqlmw

import "database/sql/driver"

// driver wraps a sql.Driver with an interceptor.
type wrappedDriver struct {
	intr   Interceptor
	parent driver.Driver
}

// Compile time validation that our types implement the expected interfaces
var (
	_ driver.Driver        = wrappedDriver{}
	_ driver.DriverContext = wrappedDriver{}
)

// WrapDriver wraps the passed driver and return a new SQL driver that has all of
// its calls intercepted by the supplied Interceptor object.
func WrapDriver(driver driver.Driver, intr Interceptor) driver.Driver {
	return wrappedDriver{parent: driver, intr: intr}
}

// Open implements the database/sql/driver.Driver interface for WrappedDriver.
func (d wrappedDriver) Open(name string) (driver.Conn, error) {
	conn, err := d.parent.Open(name)
	if err != nil {
		return nil, err
	}

	return wrappedConn{intr: d.intr, parent: conn}, nil
}

func (d wrappedDriver) OpenConnector(name string) (driver.Connector, error) {
	driver, ok := d.parent.(driver.DriverContext)
	if !ok {
		return wrappedConnector{
			parent:    dsnConnector{dsn: name, driver: d.parent},
			driverRef: &d,
		}, nil
	}
	conn, err := driver.OpenConnector(name)
	if err != nil {
		return nil, err
	}

	return wrappedConnector{parent: conn, driverRef: &d}, nil
}
