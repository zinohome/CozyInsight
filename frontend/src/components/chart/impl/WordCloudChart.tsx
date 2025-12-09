import React from 'react';
import { WordCloud } from '@ant-design/charts';
import type { WordCloudConfig } from '@ant-design/charts';

interface WordCloudChartProps {
    data: any[];
    config: {
        wordField: string;
        weightField: string;
        title?: string;
    };
}

const WordCloudChart: React.FC<WordCloudChartProps> = ({ data, config }) => {
    const chartConfig: WordCloudConfig = {
        data,
        wordField: config.wordField,
        weightField: config.weightField,
        colorField: config.wordField,
        wordStyle: {
            fontFamily: 'Verdana',
            fontSize: [8, 32],
            rotation: [0, 90],
        },
        random: () => 0.5,
        tooltip: {
            showTitle: true,
            formatter: (datum: any) => {
                return {
                    name: datum[config.wordField],
                    value: datum[config.weightField],
                };
            },
        },
    };

    return <WordCloud {...chartConfig} />;
};

export default WordCloudChart;
