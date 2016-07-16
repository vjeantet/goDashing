<?php

$arr = array(
	"current" => rand(20,500),
	"last" => rand(20,500),
	);

echo json_encode($arr);