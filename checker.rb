# frozen_string_literal: true

require 'yaml'
require './checker_services.rb'

services     = YAML.load_file('services.yaml')
service_list = []
thread_list  = []
results      = {}

services.each do |server, options|
  options['services'].each do |service, parameters|
    s =
      case parameters['type']
      when 'http'
        CheckHTTP.new(server, service, options['address'], parameters)
      when 'tcp'
        CheckTCP.new(server, service, options['address'], parameters)
      else
        CheckService.new(server, service, options['address'], parameters)
      end

    service_list.push(s)
  end
end

service_list.each do |check|
  thread_list.push(Thread.new { check.run })
end

thread_list.each(&:join)

service_list.each do |check|
  results[check.server] = {} unless results.key?(check.server)

  results[check.server][check.service] = {}
  results[check.server][check.service]['status']       = check.status
  results[check.server][check.service]['lastCheck']    = check.check_time
  results[check.server][check.service]['responseTime'] = check.response_time
end

File.open('status.yaml', 'w') do |f|
  f.write results.to_yaml
end
