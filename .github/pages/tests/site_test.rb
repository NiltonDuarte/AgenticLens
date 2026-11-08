# Exercise the production Jekyll entrypoint with temporary Markdown repositories.
require "fileutils"
require "minitest/autorun"
require "nokogiri"
require "open3"
require "pathname"
require "rbconfig"
require "tmpdir"
require "uri"

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
    @project.join(".agenticlens-config.yaml").write("workspaces:\n  dir: #{workspace.inspect}\n") if config
    plan = @project.join(workspace, "plan")
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

  def build(base_path: "/FixtureRepo", output: @output, process_options: {})
    Open3.capture3(
      RbConfig.ruby, BUILDER, "--project", @project.to_s,
      "--output", output.to_s, "--repository", "fixture-owner/FixtureRepo",
      "--revision", "0123456789abcdef", "--base-path", base_path,
      **process_options
    )
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
    refute @output.join("index.html").exist?, "The site root must not replace the /roadmap/ page"
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

  def test_default_workspace_without_configuration_or_workspace_key
    fixture(workspace: "workspaces", config: false)
    [nil, "project_name: Fixture\n", "workspaces: {}\n"].each do |config|
      @project.join(".agenticlens-config.yaml").write(config) if config
      assert_build(base_path: "")
      roadmap = document(@output.join("roadmap/index.html"))
      assert_includes links(roadmap), "/roadmap/to-do/TASK-Z%20sample.html#details"
    end
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
    content = "---\npermalink: /\n---\n# Unexpected root page\n"
    @ticket.write(content)
    stdout, stderr, status = build
    refute status.success?, "An explicit permalink must not escape /roadmap/"
    assert_includes stdout + stderr, "Jekyll output must stay under roadmap/"
    refute @output.join("index.html").exist?
    after = @output.glob("**/*", File::FNM_DOTMATCH).select(&:file?).to_h do |path|
      [path.relative_path_from(@output), path.binread]
    end
    assert_equal original_artifact, after, "Rejected output must leave the previous artifact intact"
    assert_equal content, @ticket.read
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
