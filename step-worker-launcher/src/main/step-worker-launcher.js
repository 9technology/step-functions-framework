var child_process = require('child_process');
console.log('Loading function');
var aws = require('aws-sdk');

exports.handler = function(event, context) {
    console.log('Received event:', JSON.stringify(event, null, 2));
    // Get the object from the event and show its content type
    var project = event.project;
    var env = event.env;
    var region = event.region;
    var mode = event.mode;
    var configbucket = event.configbucket;
    var configbucketkey = event.configbucketkey;
    var executionarn = event.executionarn;
    var result = event.result;

    console.log('Params', project, env, region, mode, configbucket, configbucketkey, executionarn, result);
    var proc = child_process.spawn('./stepworkerlauncher', [ project, env, region, mode, configbucket, configbucketkey, executionarn, result ], [ JSON.stringify(event) ], { stdio: 'inherit' });

    proc.stdout.on('data', function(data) {
      console.log('stdout: ' + data);
      //Here is where the output goes
    });
    proc.stderr.on('data', function(data) {
      console.log('stderr: ' + data);
      //Here is where the error output goes
    });
    proc.on('close', function(code) {
      console.log('closing code: ' + code);
      //Here you can get the exit code of the script
    });

    console.log("running lambda function")
    proc.on('close', function(code) {

      if(code !== 0) {
        return context.done(new Error("Process exited with non-zero status code:"+code));
      }

      context.done(console.log("Done running lambda function"));
    });
};
