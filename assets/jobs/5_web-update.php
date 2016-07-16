<?php
//Web API Update
define("BASEURL", $argv[1]) ;
define("TOKEN", $argv[2]) ;

function postEvent($WID, $data) 
{
	$data["auth_token"] = TOKEN ;
	$options = array(
	  'http' => array(
		'method'  => 'POST',
		'content' => json_encode( $data ),
		'header'=>  "Content-Type: application/json\r\n" .
					"Accept: application/json\r\n"
		)
	);

	$context  = stream_context_create( $options );
	$result = file_get_contents( BASEURL."/widgets/".$WID, false, $context );
}

postEvent("synergy",array("value"=>rand(0,100))) ;

echo json_encode(array("dummy"=>""));