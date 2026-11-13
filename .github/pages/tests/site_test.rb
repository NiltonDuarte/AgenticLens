# Exercise the production Jekyll entrypoint with temporary Markdown repositories.
require "fileutils"
require "minitest/autorun"
require "nokogiri"
require "open3"
require "pathname"
require "rbconfig"
require "tmpdir"
require "uri"
require "yaml"

class TicketsSiteTest < Minitest::Test
  BUILDER = File.expand_path("../build.rb", __dir__)

  def setup
    @temporary = Dir.mktmpdir("tickets-site-")
    @project = Pathname(@temporary).join("repository with spaces")
    @project.mkpath
    @output = @project.join("dist/tickets-site")
  end

  def teardown
    FileUtils.remove_entry(@temporary)
  end

  def fixture(workspace: "custom tickets", config: true)
    @project.join(".agenticlens-config.yaml").write("workspaces:\n  roots: []\n") if config
    @roadmap = @project.join(workspace, "plan")
    plan = @roadmap
    plan.join("to-do").mkpath
    plan.join("done").mkpath
    plan.join("TICKETS.md").write(<<~MARKDOWN)
      # Fixture roadmap

      ## Ready

      - **TASK-Z**: Current ticket
        - **Status**: Ready
        - **File**: [TASK-Z](to-do/TASK-Z%20sample.md#details)

      - **TASK-OLD**: Previous ticket
        - **Status**: Completed
        - **File**: [TASK-OLD](done/TASK-OLD%20archived.md)

      [API](to-do/TASK-Z%20sample.md#api--ui)

      [Repeated heading](to-do/TASK-Z%20sample.md#details-1)

      ## Blocked

      TASK-Z blocks TASK-NEXT.
    MARKDOWN
    @ticket = plan.join("to-do/TASK-Z sample.md")
    @ticket.write(<<~MARKDOWN)
      # TASK-Z

      ## Details

      **Fixture requirement**

      [Roadmap](../TICKETS.md#ready)

      [Supporting spec](../../PO-fixture/SPEC.md#context)

      [Sibling](../done/TASK-OLD%20archived.md)

      [Anchor](#details)

      [API](#api--ui)

      [Repeated heading](#details-1)

      [External](https://example.org/docs?q=value#section)

      ## API & UI

      API requirement.

      ## Details

      Second section.
    MARKDOWN
    plan.join("done/TASK-OLD archived.md").write("# Completed fixture\n")
    supporting = @project.join(workspace, "PO-fixture")
    supporting.mkpath
    supporting.join("SPEC.md").write("# Context\n")
    supporting.join("agentic_metrics.json").write('{"private_capture":"UNPUBLISHED_CAPTURE_SENTINEL"}')
    plan
  end

  def build(base_path: "/FixtureRepo", output: @output, process_options: {}, roadmap: @roadmap)
    arguments = ["--project", @project.to_s]
    arguments.concat(["--roadmap", roadmap.to_s]) unless roadmap.nil?
    Open3.capture3(
      RbConfig.ruby, (@builder || BUILDER).to_s, *arguments,
      "--output", output.to_s, "--repository", "fixture-owner/FixtureRepo",
      "--revision", "0123456789abcdef", "--base-path", base_path,
      **process_options
    )
  end

  def pages_configuration(roadmap:)
    unless @builder
      pages = Pathname(@temporary).join("Pages builder with spaces")
      FileUtils.cp_r(File.dirname(BUILDER), pages)
      @builder = pages.join("build.rb")
    end
    path = @builder.dirname.join("_config.yml")
    settings = YAML.safe_load_file(path)
    roadmap.nil? ? settings.delete("roadmap_path") : settings["roadmap_path"] = roadmap
    path.write(settings.to_yaml)
  end

  def assert_build(**arguments)
    stdout, stderr, status = build(**arguments)
    assert status.success?, stdout + stderr
  end

  def document(path)
    Nokogiri::HTML(path.read)
  end

  def links(document)
    document.css("a[href]").map { |node| node["href"] }
  end

  def test_custom_root_preserves_roadmap_structure_links_and_source_boundaries
    fixture
    originals = @project.glob("**/*", File::FNM_DOTMATCH).select(&:file?).to_h do |path|
      [path.relative_path_from(@project), path.binread]
    end
    assert_build
    welcome = document(@output.join("index.html"))
    assert_includes welcome.at_css("title").text, "AgenticLens"
    assert_match(/local.*telemetry|telemetry.*local/im, welcome.at_css("main").text)
    assert_includes links(welcome.at_css("main")), "/FixtureRepo/roadmap/"
    assert welcome.at_css("meta[name='viewport']"), "The welcome page must support narrow screens"
    roadmap = document(@output.join("roadmap/index.html"))
    %w[Ready Blocked].each { |heading| assert_includes roadmap.css("h2").map(&:text), heading }
    assert_includes roadmap.text, "Fixture roadmap"
    assert_includes links(roadmap), "/FixtureRepo/roadmap/to-do/TASK-Z%20sample.html#details"

    parents = roadmap.css("main > ul > li")
    assert_equal 2, parents.length, "Metadata must remain nested under its owning ticket"
    parents.zip(["TASK-Z", "TASK-OLD"], ["Ready", "Completed"]).each do |parent, ticket, status|
      label = parent.children.reject { |node| node.name == "ul" }.map(&:text).join
      assert_includes label, ticket, "The parent ticket label must remain visible outside metadata"
      children = parent.css("> ul > li").map { |node| node.text.split.join(" ") }
      assert_equal ["Status: #{status}", "File: #{ticket}"], children
    end

    ticket = document(@output.join("roadmap/to-do/TASK-Z sample.html"))
    assert_includes links(ticket), "/FixtureRepo/roadmap/"
    assert_includes links(ticket), "/FixtureRepo/roadmap/#ready"
    assert_includes links(ticket), "/FixtureRepo/roadmap/done/TASK-OLD%20archived.html"
    assert_includes links(ticket), "https://github.com/fixture-owner/FixtureRepo/blob/0123456789abcdef/custom%20tickets/PO-fixture/SPEC.md#context"
    assert_includes links(ticket), "https://example.org/docs?q=value#section"
    %w[api--ui details details-1].each { |fragment| assert ticket.at_css("[id='#{fragment}']"), fragment }
    %w[api--ui details-1].each do |fragment|
      assert_includes links(roadmap), "/FixtureRepo/roadmap/to-do/TASK-Z%20sample.html##{fragment}"
      assert_includes links(ticket), "##{fragment}"
    end
    assert_includes links(ticket), "#details"
    assert ticket.at_css("meta[name='viewport']"), "Narrow screens require the viewport declaration"
    assert_equal "Fixture requirement", ticket.at_css("strong").text
    originals.each { |relative, before| assert_equal before, @project.join(relative).binread }
    assert_empty @output.glob("**/*.md")
    assert_empty @output.glob("**/*.json")
    @output.glob("**/*", File::FNM_DOTMATCH).select(&:file?).each do |path|
      refute_includes path.binread, "UNPUBLISHED_CAPTURE_SENTINEL"
    end
  end

  def test_presentation_assets_fragments_and_accessible_structure_at_project_host_and_preview_roots
    fixture
    ["/FixtureRepo", "", "/FixtureRepo/preview/24"].each do |base_path|
      assert_build(base_path: base_path)
      presentation = document(@output.join("index.html"))
      assert_equal "en", presentation.at_css("html")["lang"]
      assert_equal 1, presentation.css("main").length
      assert_equal 1, presentation.css("h1").length
      assert presentation.at_css("header nav"), "Visitors need a primary navigation landmark"
      skip = presentation.at_css("body > a[href='#main']")
      assert skip, "Keyboard visitors need a skip link before the navigation"
      assert presentation.at_css("main#main"), "The skip link must reach the content landmark"
      assert_empty presentation.css("script, iframe, form"), "The presentation works without executable services or visitor data collection"
      presentation.css("[tabindex]").each do |node|
        assert_operator node["tabindex"].to_i, :<=, 0, "Natural keyboard order must not be overridden"
      end
      ids = presentation.css("[id]").map { |node| node["id"] }
      assert_equal ids.uniq, ids, "Duplicate IDs make deep-link destinations ambiguous"
      heading_levels = presentation.css("h1, h2, h3, h4, h5, h6").map { |node| node.name[1].to_i }
      heading_levels.each_cons(2) { |previous, current| assert_operator current, :<=, previous + 1 }

      presentation.css("a[href], link[href], img[src]").each do |node|
        target = node[node.name == "img" ? "src" : "href"]
        uri = URI.parse(target)
        next if uri.host || uri.scheme
        unless uri.path.empty?
          assert uri.path.start_with?("#{base_path}/"), "Local links must respect project hosting: #{target}"
        end
        relative = uri.path.delete_prefix(base_path).delete_prefix("/")
        relative = "index.html" if relative.empty?
        relative += "index.html" if relative.end_with?("/")
        destination = @output.join(URI::DEFAULT_PARSER.unescape(relative))
        assert destination.file?, "Missing local page or asset: #{target}"
        if uri.fragment && !uri.fragment.empty?
          assert document(destination).at_css("[id='#{uri.fragment}']"), "Missing fragment: #{target}"
        end
      end
      illustrations = presentation.css("figure img[src$='.svg']")
      assert_operator illustrations.length, :>=, 4, "Product and architecture explanations need captioned visuals"
      illustrations.each do |node|
        refute_empty node["alt"].to_s.strip, "Every explanatory image needs a text alternative"
        figure = node.ancestors("figure").first
        refute_empty figure.at_css("figcaption")&.text.to_s.strip, "Visuals need visible captions"
        asset = @output.join(node["src"].delete_prefix(base_path).delete_prefix("/"))
        svg = Nokogiri::XML(asset.read) { |config| config.strict.nonet }
        assert_equal "svg", svg.root.name
        assert svg.at_css("title"), "SVG explanations need an accessible title"
        assert_empty svg.css("script, foreignObject, image"), "Illustrations must be self-contained, not embedded private captures"
      end
      assert presentation.at_css("link[rel='stylesheet'][href^='#{base_path}/presentation/']"), "The trusted stylesheet must be published beneath the project path"
      roadmap = document(@output.join("roadmap/index.html"))
      assert_includes links(roadmap), "#{base_path}/roadmap/to-do/TASK-Z%20sample.html#details"
      ticket = document(@output.join("roadmap/to-do/TASK-Z sample.html"))
      assert_includes links(ticket), "#{base_path}/roadmap/"
      assert_includes links(ticket), "#{base_path}/roadmap/done/TASK-OLD%20archived.html"
      assert_includes links(ticket), "https://github.com/fixture-owner/FixtureRepo/blob/0123456789abcdef/custom%20tickets/PO-fixture/SPEC.md#context"
    end
  end

  def test_explicit_roadmap_is_independent_of_telemetry_configuration
    fixture(workspace: "workspaces", config: false)
    pages_configuration(roadmap: "missing configured roadmap")
    [nil, "project_name: Fixture\n", "workspaces:\n  roots: []\n"].each do |config|
      @project.join(".agenticlens-config.yaml").write(config) if config
      assert_build(base_path: "")
      welcome = document(@output.join("index.html"))
      assert_includes links(welcome.at_css("main")), "/roadmap/"
      roadmap = document(@output.join("roadmap/index.html"))
      assert_includes links(roadmap), "/roadmap/to-do/TASK-Z%20sample.html#details"
    end
  end

  def test_missing_roadmap_requires_explicit_selection
    fixture
    @project.join(".agenticlens-config.yaml").write("workspaces:\n  dir: custom tickets\n  roots:\n    - path: custom tickets/plan\n      discovery:\n        max_depth: 0\n        max_directories: 300\n")
    [nil, "", 42].each do |roadmap|
      pages_configuration(roadmap: roadmap)
      stdout, stderr, status = build(roadmap: nil)
      refute status.success?
      assert_includes stdout + stderr, "--roadmap PATH or Pages roadmap_path is required"
      refute @output.exist?
    end
  end

  def test_configured_roadmap_supports_caller_without_flag_and_ignores_telemetry_roots
    fixture(config: false)
    pages_configuration(roadmap: @roadmap.relative_path_from(@project).to_s)
    [nil, "invalid: [\n", "workspaces:\n  dir: nonexistent\n  roots:\n    - path: other-telemetry\n"].each do |config|
      @project.join(".agenticlens-config.yaml").write(config) if config
      assert_build(roadmap: nil)
      roadmap = document(@output.join("roadmap/index.html"))
      assert_includes links(roadmap), "/FixtureRepo/roadmap/to-do/TASK-Z%20sample.html#details"
      assert @output.join("roadmap/to-do/TASK-Z sample.html").file?
    end
    stdout, stderr, status = build(roadmap: "")
    refute status.success?, "An explicitly blank CLI selection must not use configured input"
    assert_includes stdout + stderr, "--roadmap PATH or Pages roadmap_path is required"
    assert @output.join("roadmap/to-do/TASK-Z sample.html").file?
  end

  def test_configured_roadmap_cannot_escape_repository_and_preserves_published_output
    fixture
    assert_build
    original = @output.glob("**/*", File::FNM_DOTMATCH).select(&:file?).to_h do |path|
      [path.relative_path_from(@output), path.binread]
    end
    outside = Pathname(@temporary).join("outside repository")
    outside.mkpath
    outside.join("TICKETS.md").write("# Outside source\n")
    @project.join("roadmap-link").make_symlink(outside)
    [outside.to_s, "roadmap-link"].each do |roadmap|
      pages_configuration(roadmap: roadmap)
      stdout, stderr, status = build(roadmap: nil)
      refute status.success?, "Configured roadmap escaped the project: #{roadmap}"
      assert_includes stdout + stderr, "Plan must be inside the repository"
      after = @output.glob("**/*", File::FNM_DOTMATCH).select(&:file?).to_h do |path|
        [path.relative_path_from(@output), path.binread]
      end
      assert_equal original, after
    end
    assert_equal "# Outside source\n", outside.join("TICKETS.md").read
  end

  def test_repeated_build_removes_stale_pages_and_renders_source_updates
    plan = fixture
    assert_build
    archived = @output.join("roadmap/done/TASK-OLD archived.html")
    assert archived.exist?
    plan.join("done/TASK-OLD archived.md").delete
    plan.join("TICKETS.md").write("# Roadmap\n\n[Current](to-do/TASK-Z%20sample.md)\n")
    @ticket.write("# TASK-Z\n\nUpdated requirement.\n")
    assert_build
    refute archived.exist?
    assert_includes document(@output.join("roadmap/to-do/TASK-Z sample.html")).text, "Updated requirement."
  end

  def test_invalid_inputs_and_outputs_preserve_source_and_unowned_files
    plan = fixture
    roadmap = plan.join("TICKETS.md")
    original = roadmap.binread
    [@project, plan, plan.join("generated")].each do |output|
      _stdout, _stderr, status = build(output: output)
      refute status.success?, "Unsafe output accepted: #{output}"
      assert_equal original, roadmap.binread
    end
    @output.mkpath
    unowned = @output.join("user-owned.txt")
    unowned.write("retain this file")
    _stdout, _stderr, status = build
    refute status.success?
    assert_equal "retain this file", unowned.read
    roadmap.delete
    stdout, stderr, status = build(output: @project.join("dist/missing-input"))
    refute status.success?
    assert_includes stdout + stderr, "TICKETS.md"
  end

  def test_encoded_supporting_filename_does_not_become_query_or_fragment
    fixture
    supporting = @project.join("custom tickets/PO-fixture/SPEC#why?100%.md")
    supporting.write("# Context\n")
    @ticket.write("# TASK-Z\n\n[Supporting](../../PO-fixture/SPEC%23why%3F100%25.md?mode=raw#context)\n")
    assert_build
    ticket = document(@output.join("roadmap/to-do/TASK-Z sample.html"))
    expected = "https://github.com/fixture-owner/FixtureRepo/blob/0123456789abcdef/custom%20tickets/PO-fixture/SPEC%23why%3F100%25.md?mode=raw#context"
    assert_includes links(ticket), expected
    parsed = URI.parse(ticket.at_css("main a")["href"])
    assert_equal "/fixture-owner/FixtureRepo/blob/0123456789abcdef/custom%20tickets/PO-fixture/SPEC%23why%3F100%25.md", parsed.path
    assert_equal "mode=raw", parsed.query
    assert_equal "context", parsed.fragment
    assert_equal "# Context\n", supporting.read
  end

  def test_front_matter_cannot_publish_outside_roadmap_or_replace_previous_artifact
    fixture
    assert_build
    original_artifact = @output.glob("**/*", File::FNM_DOTMATCH).select(&:file?).to_h do |path|
      [path.relative_path_from(@output), path.binread]
    end
    assert original_artifact.key?(Pathname("index.html")), "The preserved artifact must include the welcome page"
    ["/", "/index.html", "/roadmap/%2e%2e/index.html", "/presentation/style.css", "/presentation/private.html"].each do |permalink|
      content = "---\npermalink: #{permalink}\n---\n# Unexpected root page\n"
      @ticket.write(content)
      stdout, stderr, status = build
      refute status.success?, "Ticket permalink #{permalink.inspect} must not escape /roadmap/"
      assert_includes stdout + stderr, "roadmap/"
      after = @output.glob("**/*", File::FNM_DOTMATCH).select(&:file?).to_h do |path|
        [path.relative_path_from(@output), path.binread]
      end
      assert_equal original_artifact, after, "Rejected output must preserve the welcome page and roadmap artifact"
      assert_equal content, @ticket.read
    end
  end

  def test_permission_failure_rejects_rebuild_without_nested_or_stale_success
    fixture
    assert_build
    locked = @output.join("locked")
    locked.mkpath
    locked.join("stale.txt").write("old artifact")
    @ticket.write("# TASK-Z\n\nChanged source for rebuild.\n")
    source_before = @ticket.binread
    process_options = {}
    if Process.uid.zero?
      # Root bypasses mode bits; make this real filesystem fixture accessible to
      # an unprivileged child so the denied removal is exercised in Docker too.
      File.chmod(0o755, @temporary)
      FileUtils.chown_R(65_534, 65_534, @temporary)
      process_options = {uid: 65_534, gid: 65_534}
    end
    File.chmod(0o555, locked)
    stdout, stderr, status = build(process_options: process_options)
    refute status.success?, "An unreadable cleanup must not report a successful rebuild"
    assert_match(/EACCES|Permission denied/, stdout + stderr)
    refute @output.join("_site").exist?, "Failed replacement must not nest the new artifact in stale output"
    assert_equal source_before, @ticket.binread
    assert_equal "old artifact", locked.join("stale.txt").read
  ensure
    File.chmod(0o755, locked) if locked&.directory?
  end
end
