<?php
	echo "Hello";
	isset($_POST["out"]) ? $rx = $_POST["out"] : $rx = -1;
	if($rx == -1)
		die("Error");

	$object = json_decode($rx);
	$server = "localhost";
	$user = "s5535512104";
	$password = "4128909825";
	$db = "s5535512104"
	// Create connection
	$conn = new mysqli($servername, $username, $password, $dbname);
	// Check connection
	if ($conn->connect_error) {
    	die();
	} 
/*
	$sql = "INSERT INTO block(lat, long) VALUES($object['lat'], $object['long']);"
	if ($conn->query($sql) === TRUE) {
    	echo json_encode($arrayName = $arrayName = array('status' => TRUE ););
	} else {
    	echo json_encode($arrayName = array('status' => FALSE, 'data' => $con->error));
	}

	$con->close();*/
?>