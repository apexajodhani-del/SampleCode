package main

func pageHTML() string {
	return `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Shape Drawer</title>
	</head>
	<body>
		<h2>Shape Drawer</h2>
		<form method="GET">
			<input type="number" name="sides" min="1" />
			<button type="submit">Draw</button>
		</form>
	</body>
	</html>
	`
}
