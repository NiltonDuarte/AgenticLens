# frozen_string_literal: true

require "fileutils"
require "optparse"
require "pathname"
require "rbconfig"
require "tmpdir"
require "yaml"

options = {project: ".", output: "dist/tickets-site", base_path: ""}
OptionParser.new do |parser|
  parser.on("--project PATH") { |value| options[:project] = value }
  parser.on("--output PATH") { |value| options[:output] = value }
  parser.on("--repository OWNER/REPO") { |value| options[:repository] = value }
  parser.on("--revision SHA") { |value| options[:revision] = value }
  parser.on("--base-path PATH") { |value| options[:base_path] = value }
end.parse!
raise "--repository OWNER/REPO is required" unless options[:repository].to_s.match?(%r{\A[\w.-]+/[\w.-]+\z})
raise "--revision is required" if options[:revision].to_s.strip.empty?
raise "--base-path must be an absolute URL path" unless options[:base_path].empty? || options[:base_path].match?(%r{\A/(?!/)[^?#]*\z})

project = Pathname.new(options[:project]).realpath
config_path = project.join(".agenticlens-config.yaml")
config = config_path.file? ? YAML.safe_load_file(config_path) || {} : {}
workspace = config.fetch("workspaces", {}).fetch("dir", "workspaces")
raise "workspaces.dir must be a nonempty path" unless workspace.is_a?(String) && !workspace.strip.empty?
plan = project.join(workspace, "plan").realpath
raise "Plan must be inside the repository" unless plan.to_s.start_with?("#{project}/")
raise "Roadmap not found: #{plan}/TICKETS.md" unless plan.join("TICKETS.md").file?
output = project.join(options[:output]).expand_path
# This entrypoint owns only generated distribution output, never arbitrary sources.
raise "Output must be under the repository's dist directory" unless output.to_s.start_with?("#{project}/dist/")
if output == plan || plan.to_s.start_with?("#{output}/") || output.to_s.start_with?("#{plan}/")
  raise "Output must be separate from the plan"
end
raise "Output must not be a symlink" if output.ascend.any?(&:symlink?)
if output.exist? && (!output.directory? || (!output.children.empty? && !output.join(".tickets-site-output").file?))
  raise "Refusing to replace unmanaged output"
end

Dir.mktmpdir("ticket-jekyll-") do |staging|
  source = Pathname.new(staging)
  %w[_layouts _plugins roadmap].each { |name| FileUtils.cp_r(File.join(__dir__, name), source.join(name)) }
  Dir.glob(plan.join("**/*.md")).each do |path|
    file = Pathname.new(path)
    raise "Plan source escapes through a symlink: #{file}" unless file.realpath.to_s.start_with?("#{plan}/")
    destination = source.join("roadmap", file.relative_path_from(plan))
    FileUtils.mkdir_p(destination.dirname)
    FileUtils.cp(file, destination)
  end
  settings = YAML.safe_load_file(File.join(__dir__, "_config.yml")).merge(
    "baseurl" => options[:base_path].delete_suffix("/"),
    "ticket_project" => project.to_s, "ticket_plan" => plan.relative_path_from(project).to_s,
    "ticket_repository" => options[:repository], "ticket_revision" => options[:revision]
  )
  source.join("_config.yml").write(settings.to_yaml)
  # Render into temporary output first so invalid input leaves the previous artifact intact.
  built = source.join("_site")
  success = system(RbConfig.ruby, "-S", "jekyll", "build", "--source", source.to_s, "--destination", built.to_s)
  raise "Jekyll build failed" unless success
  raise "Jekyll did not render the roadmap" unless built.join("roadmap/index.html").file?
  unless built.children.all? { |entry| entry.basename.to_s == "roadmap" && entry.directory? }
    raise "Jekyll output must stay under roadmap/"
  end
  built.join(".tickets-site-output").write("Generated ticket website\n")
  FileUtils.mkdir_p(output.dirname)
  FileUtils.remove_entry(output) if output.exist?
  begin
    File.rename(built, output)
  rescue Errno::EXDEV
    FileUtils.copy_entry(built, output)
  end
end
