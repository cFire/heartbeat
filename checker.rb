require 'thread'
require 'yaml'
require './checkerServices.rb'


services    = YAML.load_file('services.yaml')
serviceList = []
threadList  = []
results     = {}

services.each do |server, options|
  options['services'].each do |service, parameters|

    s = case parameters['type']
    when 'http'
      CheckHTTP.new(server, service, options['address'], parameters)
    when 'tcp'
      CheckTCP.new(server, service, options['address'], parameters)
    else
      CheckService.new(server, service, options['address'], parameters)
    end

    serviceList.push(s)
  end
end

serviceList.each do |check|
  threadList.push(Thread.new{ check.run })
end

threadList.each do |thread|
  thread.join
end

serviceList.each do |check|
  if(not results.has_key?(check.server))
    results[check.server] = {}
  end

  results[check.server][check.service] = {}
  results[check.server][check.service]['status']       = check.status
  results[check.server][check.service]['lastCheck']    = check.checkTime
  results[check.server][check.service]['responseTime'] = check.responseTime
end

File.open('status.yaml', 'w') do |f|
  f.write results.to_yaml
end