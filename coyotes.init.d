#! /bin/sh

### BEGIN INIT INFO
# Provides:          coyotes
# Required-Start:    $remote_fs $network
# Required-Stop:     $remote_fs $network
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: starts coyotes
# Description:       starts the coyotes daemon
### END INIT INFO

coyotes_BIN=/usr/bin/coyotes
coyotes_PID=/tmp/coyotes.pid
coyotes_USER=www
coyotes_LOG=/home/data/logs/task-runner.log

# 在这里添加启动参数
coyotes_opts="-debug=true -pidfile=$coyotes_PID"

wait_for_pid () {
	try=0

	while test $try -lt 35 ; do

		case "$1" in
			'created')
			if [ -f "$2" ] ; then
				try=''
				break
			fi
			;;

			'removed')
			if [ ! -f "$2" ] ; then
				try=''
				break
			fi
			;;
		esac

		echo -n .
		try=`expr $try + 1`
		sleep 1

	done

}

case "$1" in
	start)
		echo -n "Starting coyotes "

		sudo -u $coyotes_USER bash -c "$coyotes_BIN -daemonize=true $coyotes_opts >> $coyotes_LOG 2>&1"

		if [ "$?" != 0 ] ; then
			echo " failed"
			exit 1
		fi

		wait_for_pid created $coyotes_PID

		if [ -n "$try" ] ; then
			echo " failed"
			exit 1
		else
			echo " done"
		fi
	;;

	stop)
		echo -n "Gracefully shutting down coyotes "

		if [ ! -r $coyotes_PID ] ; then
			echo "warning, no pid file found - coyotes is not running ?"
			exit 1
		fi

		kill -USR2 `cat $coyotes_PID`

		wait_for_pid removed $coyotes_PID

		if [ -n "$try" ] ; then
			echo " failed. Use force-quit"
			exit 1
		else
			echo " done"
		fi
	;;

	status)
		if [ ! -r $coyotes_PID ] ; then
			echo "coyotes is stopped"
			exit 0
		fi

		PID=`cat $coyotes_PID`
		if ps -p $PID | grep -q $PID; then
			echo "coyotes (pid $PID) is running..."
		else
			echo "coyotes dead but pid file exists"
		fi
	;;

	force-quit)
		echo -n "Terminating coyotes "

		if [ ! -r $coyotes_PID ] ; then
			echo "warning, no pid file found - coyotes is not running ?"
			exit 1
		fi

		kill -TERM `cat $coyotes_PID`

		wait_for_pid removed $coyotes_PID

		if [ -n "$try" ] ; then
			echo " failed"
			exit 1
		else
			echo " done"
		fi
	;;

	restart)
		$0 stop
		$0 start
	;;

	*)
		echo "Usage: $0 {start|stop|force-quit|restart}"
		exit 1
	;;

esac