#!/bin/sh

# goji-invited-user-signup-example - this script starts and stops the goji-invited-user-signup-example daemon
#
# chkconfig: 345 85 15
# description: goji-invited-user-signup-example
# processname: goji-invited-user-signup-example

# Source function library.
. /etc/init.d/functions

# Source networking configuration.
. /etc/sysconfig/network
 
# Check that networking is up.
[ "$NETWORKING" = "no" ] && exit 0

user="nobody"
approot="/path/to/goji-invited-user-signup-example"
prog="goji-invited-user-signup-example"
cmd="cd $approot; nohup ./${prog}"
sysconfig="/etc/sysconfig/${prog}"
lockfile="/var/lock/subsys/${prog}"

[ -f $sysconfig ] && . $sysconfig

start() {
    [ -x $cmd ] || exit 5
    echo -n $"Starting $prog: "
    daemon --user=$user $cmd >/dev/null 2>&1 &
    retval=$?
    echo
    [ $retval -eq 0 ] && touch $lockfile
    return $retval
}
	
stop() {
    echo -n $"Stopping $prog: "
    killproc $prog
    retval=$?
    echo
    [ $retval -eq 0 ] && rm -f $lockfile
    return $retval
}

restart() {
    stop
    sleep 1
    start
}

rh_status() {
    status $prog
}
 
rh_status_q() {
    rh_status >/dev/null 2>&1
}

case "$1" in
  start)
      rh_status_q && exit 0
      $1
      ;;
  stop)
      rh_status_q || exit 0
      $1
      ;;
  restart)
      $1
      ;;
  status)
      rh_status
      ;;

  *)
      echo $"Usage: $0 {start|stop|status|restart}"
      exit 2
esac

exit $?
