import React from "react";
import {Button, Col, Input, Modal, Rate, Row, Table, Tag} from 'antd';
import {CheckCircleOutlined, SyncOutlined, CloseCircleOutlined, ExclamationCircleOutlined, MinusCircleOutlined} from '@ant-design/icons';
import * as StudentBackend from "./backend/StudentBackend";
import * as ProgramBackend from "./backend/ProgramBackend";
import * as ReportBackend from "./backend/ReportBackend";
import * as RoundBackend from "./backend/RoundBackend";
import moment from "moment";
import * as Setting from "./Setting";
import {CSVLink} from "react-csv";
import ReactMarkdown from "react-markdown";

class RankingPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      programName: props.match.params.programName,
      students: null,
      reports: null,
      program: null,
      columns: this.getColumns(),
      reportVisible: false,
      report: null,
    };
  }

  getColumns() {
    return [
      {
        title: 'Name',
        dataIndex: 'realName',
        key: 'realName',
        width: '60px',
        render: (text, record, index) => {
          return (
            <a href={`/user/${record.name}`}>{text}</a>
          )
        }
      },
      {
        title: 'GitHub',
        dataIndex: 'github',
        key: 'github',
        width: '80px',
        ellipsis: true,
        render: (text, record, index) => {
          return (
            <a target="_blank" href={`https://github.com/${text}`}>{text}</a>
          )
        }
      },
      {
        title: 'Mentor',
        dataIndex: 'mentor',
        key: 'mentor',
        width: '70px',
        render: (text, record, index) => {
          return (
            <a target="_blank" href={`https://github.com/${text}`}>{text}</a>
          )
        }
      },
      {
        title: 'Score',
        dataIndex: 'score',
        key: 'score',
        width: '50px',
      },
    ];
  }

  isCurrentRound(round) {
    const now = moment();
    return moment(round.startDate) <= now && now < moment(round.endDate);
  }

  openReport(report) {
    this.setState({
      reportVisible: true,
      report: report,
    });
  }

  getTag(report) {
    if (report.text === "") {
      return (
        <Tag style={{cursor: "pointer"}} icon={<CloseCircleOutlined />} color="error">N/A</Tag>
      )
    }

    if (report.score <= 0) {
      return (
        <Tag style={{cursor: "pointer"}} icon={<MinusCircleOutlined />} color="error">{report.score}</Tag>
      )
    } else if (report.score <= 2) {
      return (
        <Tag style={{cursor: "pointer"}} icon={<ExclamationCircleOutlined />} color="warning">{report.score}</Tag>
      )
    } else if (report.score <= 4) {
      return (
        <Tag style={{cursor: "pointer"}} icon={<SyncOutlined spin />} color="processing">{report.score}</Tag>
      )
    } else {
      return (
        <Tag style={{cursor: "pointer"}} icon={<CheckCircleOutlined />} color="success">{report.score}</Tag>
      )
    }
  }

  newReport(program, round, student) {
    return {
      owner: "admin", // this.props.account.username,
      name: `report_${program.name}_${round.name}_${student.name}`,
      createdTime: moment().format(),
      program: program.name,
      round: round.name,
      student: student.name,
      text: "",
      score: 0,
    }
  }

  componentWillMount() {
    Promise.all([this.getFilteredStudents(this.state.programName), this.getFilteredReports(this.state.programName), this.getFilteredRounds(this.state.programName), this.getProgram(this.state.programName)]).then((values) => {
      let students = values[0];
      let reports = values[1];
      let rounds = values[2];
      let program = values[3];

      let roundColumns = [];
      rounds.forEach((round) => {
        roundColumns.push(
          {
            title: <a href={`/rounds/${round.name}`}>{round.name}</a>,
            dataIndex: round.name,
            key: round.name,
            width: '70px',
            // sorter: (a, b) => a.key.localeCompare(b.key),
            className: this.isCurrentRound(round) ? "alert-row" : null,
            render: (report, student, index) => {
              return (
                <a onClick={() => this.openReport.bind(this)(report)}>
                  {
                    this.getTag(report)
                  }
                </a>
              )
            }
          },
        );
      });

      let studentMap = new Map();
      students.forEach(student => {
        student.score = 0;
        studentMap.set(student.name, student);
      });
      let roundMap = new Map();
      rounds.forEach(round => {
        roundMap.set(round.name, round);

        students.forEach(student => {
          student[round.name] = this.newReport(program, round, student);
        });
      });

      reports.forEach((report) => {
        const roundName = report.round;
        const studentName = report.student;

        let student = studentMap.get(studentName);
        student[roundName] = report;
        student.score += report.score;
      });

      students.sort(function(a, b) {
        return b.score - a.score;
      });

      const columns = this.state.columns.concat(roundColumns);
      this.initCsv(students, columns);
      this.setState({
        students: students,
        reports: reports,
        columns: columns,
        program: program,
      });
    });
  }

  getFilteredStudents(programName) {
    return StudentBackend.getFilteredStudents("admin", programName)
      .then((res) => {
        return res;
      });
  }

  getFilteredReports(programName) {
    return ReportBackend.getFilteredReports("admin", programName)
      .then((res) => {
        return res;
      });
  }

  getFilteredRounds(programName) {
    return RoundBackend.getFilteredRounds("admin", programName)
      .then((res) => {
        return res;
      });
  }

  getProgram(programName) {
    return ProgramBackend.getProgram("admin", programName)
      .then((res) => {
        return res;
      });
  }

  initCsv(students, columns) {
    let data = [];
    students.forEach((student, i) => {
      let row = {};

      columns.forEach((column, i) => {
        row[column.key] = Setting.toCsv(student[column.key]);
      });

      data.push(row);
    });

    let headers = columns.map(column => {
      return {label: column.title, key: column.key};
    });
    headers = headers.slice(0, 4);

    this.setState({
      csvData: data,
      csvHeaders: headers,
    });
  }

  renderDownloadCsvButton() {
    if (this.state.csvData === null || this.state.students === null) {
      return null;
    }

    return (
      <CSVLink data={this.state.csvData} headers={this.state.csvHeaders} filename={`Ranking-${this.state.programName}.csv`}>
        <Button type="primary" size="small">Download CSV</Button>
      </CSVLink>
    )
  }

  renderTable(students) {
    return (
      <div>
        <Table columns={this.state.columns} dataSource={students} rowKey="name" size="middle" bordered pagination={{pageSize: 100}}
               title={() => (
                 <div>
                   {`"${this.state.program?.title}"`} Ranking&nbsp;&nbsp;&nbsp;&nbsp;
                   {
                     this.renderDownloadCsvButton()
                   }
                 </div>
               )}
               loading={students === null}
        />
      </div>
    );
  }

  submitReportEdit() {
    let report = Setting.deepCopy(this.state.report);
    ReportBackend.updateReport(this.state.report.owner, this.state.report.name, report)
      .then((res) => {
        if (res) {
          Setting.showMessage("success", `Successfully saved`);
          window.location.reload();
        } else {
          Setting.showMessage("error", `failed to save: server side failure`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `failed to save: ${error}`);
      });
  }

  handleReportOk() {
    this.submitReportEdit();
    this.setState({
      reportVisible: false,
    });
  }

  handleReportCancel() {
    this.setState({
      reportVisible: false,
    });
  }

  parseReportField(key, value) {
    if (["score"].includes(key)) {
      value = Setting.myParseInt(value);
    }
    return value;
  }

  updateReportField(key, value) {
    value = this.parseReportField(key, value);

    let report = Setting.deepCopy(this.state.report);
    report[key] = value;
    this.setState({
      report: report,
    });
  }

  renderReportModal() {
    if (this.state.report === null) {
      return null;
    }

    const desc = [
      '1 - Terrible: did nothing or empty weekly report',
      '2 - Bad: just relied to one or two issues, no much code contribution',
      '3 - Normal: just so so',
      '4 - Good: had made a good progress',
      '5 - Wonderful: you are a genius!'];

    return (
      <Modal
        title={`Weekly Report for ${this.state.report.round} - ${this.state.report.student}`}
        visible={this.state.reportVisible}
        onOk={this.handleReportOk.bind(this)}
        onCancel={this.handleReportCancel.bind(this)}
        okText="Save"
        width={1000}
      >
        <div>
          <ReactMarkdown
            source={this.state.report.text}
            renderers={{image: props => <img {...props} style={{maxWidth: '100%'}} alt="img" />}}
            escapeHtml={false}
          />
          <Rate tooltips={desc} value={this.state.report.score} onChange={value => {
            this.updateReportField('score', value);
          }} />
        </div>
      </Modal>
    )
  }

  render() {
    return (
      <div>
        <Row style={{width: "100%"}}>
          <Col span={1}>
          </Col>
          <Col span={22}>
            {
              this.renderTable(this.state.students)
            }
          </Col>
          <Col span={1}>
          </Col>
          {
            this.renderReportModal()
          }
        </Row>
      </div>
    );
  }
}

export default RankingPage;
