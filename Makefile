
db-diagram:
	d2 --layout=elk -t 101 d2/db.d2 d2/db.png

dev-db-diagram:
	d2 --layout=elk -w -t 101 d2/db.d2 d2/db.png