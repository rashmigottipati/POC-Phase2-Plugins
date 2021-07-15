import sys
import argparse
import json
import os

class PluginRequest(json.JSONDecoder):
  def __init__(self):
    self.command = ''
    self.args = []
    self.universe = {}

  def decode(self, s):
    dct = json.loads(s)
    if 'command' in dct:
      self.command = dct['command']
      self.args = dct['args']
      self.universe = dct['universe']
      return self
    return dct

class PluginResponse():
  def __init__(self, cmd, universe={}):
    self.command = cmd
    self.universe = universe
    self.error = False
    self.error_msg = ''

  def json(self):
    return json.dumps(self, default=lambda o: o.__dict__, sort_keys=True)

class PluginError(PluginResponse):
  def __init__(self, cmd, msg):
    super().__init__(cmd, {})
    self.error = True
    self.error_msg = msg


def parse_init_args(args_list):
  p = argparse.ArgumentParser(description='Parse "init" plugin flags')
  p.add_argument('--domain', type=str, required=False, default='my.domain', help='Project domain')
  p.add_argument('--license', type=str, required=False, default='apache2', help='Project license')
  return p.parse_args(args_list)

def run_init(req: PluginRequest):
  universe = {}

  args = parse_init_args(req.args)

  if args.__dict__['license'] == 'apache2':
    universe['LICENSE'] = 'Apache 2.0 License\n'

  universe['main.py'] = f"""
def hello_domain():
  print('Hello, {args.__dict__['domain']}!')

if __name__ == "__main__":
  hello_domain()
"""

  return PluginResponse(req.command, universe)

def parse_create_api_args(args_list):
  p = argparse.ArgumentParser(description='Parse "create api" plugin flags')
  p.add_argument('--group', type=str, required=True, help='API simple group')
  p.add_argument('--version', type=str, required=True, help='API version')
  p.add_argument('--kind', type=str, required=True, help='API kind')
  return p.parse_args(args_list)

def run_create_api(req: PluginRequest):

  universe = req.universe

  args = parse_create_api_args(req.args)

  if os.path.exists('gvk.py'):
    return PluginError(req.command, 'gvk.py must not exist')

  universe['gvk.py'] = f"""
class {args.__dict__['kind']}():

  group = '{args.__dict__['group']}'
  version = '{args.__dict__['version']}'
  kind = '{args.__dict__['kind']}'

  def __init__(self, name, namespace):
    self.name = name
    self.namespace = namespace
"""

  return PluginResponse(req.command, universe)

if __name__ == "__main__":

  req = json.loads(sys.stdin.read(), cls=PluginRequest)

  if req.command == 'init':
    res = run_init(req)
  elif req.command == 'create api':
    res = run_create_api(req)
  else:
    res = PluginError(req.command, f'plugin not supported: {req.command}')

  sys.stdout.write(res.json())
