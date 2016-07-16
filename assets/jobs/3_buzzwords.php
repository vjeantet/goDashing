<?php

$arr = array( 
	"items" => array(
		array("label"=> "item 1", "value"=> rand(0,500)),
		array("label"=> "item 2 ", "value"=> rand(0,500)),
		array("label"=> "item 3", "value"=> rand(0,500)),
		array("label"=> "item 4", "value"=> rand(0,500)),
		array("label"=> "item 5", "value"=> rand(0,500)),
		array("label"=> "item 6", "value"=> rand(0,500)),
		array("label"=> "item 7 ", "value"=> rand(0,500)),
		array("label"=> "item 8", "value"=> rand(0,500)),
		array("label"=> "item 9", "value"=> rand(0,500)),
		array("label"=> "item 10", "value"=> rand(0,500)),
		array("label"=> "item 11", "value"=> rand(0,500)),
		),

	);

echo json_encode($arr);
