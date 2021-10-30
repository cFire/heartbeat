# frozen_string_literal: true

class EndOfRead < RuntimeError
end

# Parent class that all checks are abstracted from
class CheckService
  require 'timeout'

  attr_reader :server, :service, :status, :response_time, :check_time

  def initialize(server, service, address, parameters)
    @server        = server
    @service       = service
    @address       = address
    @parameters    = parameters
    @status        = 'DOWN'
    @response_time = 0
    @check_time    = Time.now
  end

  def run
    @status        = 'DOWN'
    @response_time = 0
    @check_time    = Time.now
  end
end

# Checks for http/https responses
class CheckHTTP < CheckService
  require 'nokogiri'
  require 'open-uri'
  require 'openssl'

  def run
    start = Time.now.to_f * 1000

    uri_check

    @response_time = ((Time.now.to_f * 1000) - start).floor
    @check_time    = Time.now
  end

  def uri_builder
    uri = @parameters['ssl'] ? 'https://' : 'http://'
    uri += @parameters.key?('vhost') ? @parameters['vhost'] : @address
    uri + ":#{@parameters['port']}#{@parameters['get']}"
  end

  # rubocop:disable Security/Open
  def uri_check
    Timeout.timeout(10) do
      doc = Nokogiri::HTML open(uri_builder, redirect: false)
      @status = doc.to_s.include?(@parameters['expect']) ? 'OK' : 'WARNING'
    end
  rescue OpenURI::HTTPRedirect, OpenURI::HTTPError => e
    @status = e.to_s.include?(@parameters['expect']) ? 'OK' : 'WARNING'
  rescue StandardError => e
    $stderr.puts e
  end
  # rubocop:enable Security/Open
end

# Checks for responses on TCP sockets
class CheckTCP < CheckService
  require 'socket'
  require 'openssl'

  def run
    start = Time.now.to_f * 1000

    begin
      Timeout.timeout(10) do
        socket_check
      end
    rescue StandardError => e
      $stderr.puts e
    end

    @response_time = ((Time.now.to_f * 1000) - start).floor
    @check_time    = Time.now
  end

  def socket_open
    s = TCPSocket.open(@address, @parameters['port'])
    if @parameters['ssl']
      ssl_context = OpenSSL::SSL::SSLContext.new
      socket = OpenSSL::SSL::SSLSocket.new(s, ssl_context)
      socket.sync_close = true
      socket.connect
    else
      socket = s
    end
    socket
  end

  def socket_check(response = '')
    socket = socket_open

    Timeout.timeout(3, EndOfRead) do
      socket.puts(@parameters['send']) if @parameters.key?('send')
      while response += socket.gets
        raise EndOfRead if response.include?(@parameters['expect'])
      end
    end
  rescue EndOfRead
    @status = response.include?(@parameters['expect']) ? 'OK' : 'WARNING'
    socket.close
  end
end
